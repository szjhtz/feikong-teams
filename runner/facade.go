// Package runner 保留旧 Runner 工厂入口，实际实现位于 internal/app/agent。
package runner

import appagent "fkteams/internal/app/agent"

const (
	ModeTeam       = appagent.ModeTeam
	ModeSupervisor = appagent.ModeSupervisor
	ModeRoundtable = appagent.ModeRoundtable
	ModeCustom     = appagent.ModeCustom
	ModeDeep       = appagent.ModeDeep
)

type Cache = appagent.Cache

var NewCache = appagent.NewCache
var Resolve = appagent.Resolve
var CreateBackgroundTaskRunner = appagent.CreateBackgroundTaskRunner
var CreateAgentRunner = appagent.CreateAgentRunner
var CreateTeamRunner = appagent.CreateTeamRunner
var CreateDeepAgentsRunner = appagent.CreateDeepAgentsRunner
var CreateLoopAgentRunner = appagent.CreateLoopAgentRunner
var CreateCustomRunner = appagent.CreateCustomRunner
var PrintCustomAgentsInfo = appagent.PrintCustomAgentsInfo
var PrintLoopAgentsInfo = appagent.PrintLoopAgentsInfo
