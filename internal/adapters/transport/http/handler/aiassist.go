package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	appaiassist "fkteams/internal/app/aiassist"
	"fkteams/internal/app/config"
	appskill "fkteams/internal/app/skill"
	apptools "fkteams/internal/app/tools"

	"github.com/gin-gonic/gin"
)

const statusClientClosedRequest = 499

func GenerateAgentDraftsHandler() gin.HandlerFunc {
	return NewRuntime().GenerateAgentDraftsHandler()
}

func (rt *Runtime) GenerateAgentDraftsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req appaiassist.AgentDraftRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, http.StatusBadRequest, "invalid request: "+err.Error())
			return
		}
		ctx, cancel := context.WithTimeout(rt.withRuntimeContext(c.Request.Context()), 60*time.Second)
		defer cancel()
		enrichAgentDraftRequest(ctx, &req, config.Get())
		service, err := appaiassist.NewDefault(ctx)
		if err != nil {
			Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := service.GenerateAgents(ctx, req)
		if err != nil {
			failAIAssistError(c, err)
			return
		}
		OK(c, resp)
	}
}

func GenerateSkillDraftHandler() gin.HandlerFunc {
	return NewRuntime().GenerateSkillDraftHandler()
}

func (rt *Runtime) GenerateSkillDraftHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req appaiassist.SkillDraftRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, http.StatusBadRequest, "invalid request: "+err.Error())
			return
		}
		ctx, cancel := context.WithTimeout(rt.withRuntimeContext(c.Request.Context()), 60*time.Second)
		defer cancel()
		enrichSkillDraftRequest(&req)
		service, err := appaiassist.NewDefault(ctx)
		if err != nil {
			Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := service.GenerateSkill(ctx, req)
		if err != nil {
			failAIAssistError(c, err)
			return
		}
		OK(c, resp)
	}
}

func RewriteTextHandler() gin.HandlerFunc {
	return NewRuntime().RewriteTextHandler()
}

func enrichSkillDraftRequest(req *appaiassist.SkillDraftRequest) {
	if req == nil || len(req.ExistingSkills) > 0 {
		return
	}
	skills, err := appskill.ListLocalSkills()
	if err != nil {
		return
	}
	for _, item := range skills {
		if item.Slug != "" {
			req.ExistingSkills = append(req.ExistingSkills, item.Slug)
		}
	}
}

func (rt *Runtime) RewriteTextHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req appaiassist.RewriteTextRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, http.StatusBadRequest, "invalid request: "+err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(rt.withRuntimeContext(c.Request.Context()), 60*time.Second)
		defer cancel()
		service, err := appaiassist.NewDefault(ctx)
		if err != nil {
			Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := service.RewriteText(ctx, req)
		if err != nil {
			failAIAssistError(c, err)
			return
		}
		OK(c, resp)
	}
}

func failAIAssistError(c *gin.Context, err error) {
	if errors.Is(err, context.Canceled) || errors.Is(c.Request.Context().Err(), context.Canceled) {
		Fail(c, statusClientClosedRequest, "client closed request")
		return
	}
	Fail(c, http.StatusBadRequest, err.Error())
}

func enrichAgentDraftRequest(ctx context.Context, req *appaiassist.AgentDraftRequest, cfg *config.Config) {
	if req == nil || cfg == nil {
		return
	}
	if len(req.ExistingAgents) == 0 {
		for _, agent := range cfg.Agents.Items {
			id := agent.ID
			if id == "" {
				id = agent.Name
			}
			if id != "" {
				req.ExistingAgents = append(req.ExistingAgents, id)
			}
		}
	}
	if len(req.AvailableModels) == 0 {
		for _, model := range cfg.Models {
			if model.ID != "" {
				req.AvailableModels = append(req.AvailableModels, model.ID)
			}
		}
	}
	if len(req.AvailableTools) == 0 {
		req.AvailableTools = apptools.GetAllToolNames(ctx)
	}
	if req.DefaultModelID == "" {
		if model := cfg.ResolveDefaultModel(config.ModelUseChat); model != nil {
			req.DefaultModelID = model.ID
		}
	}
}
