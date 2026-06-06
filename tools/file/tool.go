package file

import (
	"fkteams/agentcore"
	"fmt"
)

// GetTools 获取所有文件操作工具
func (ft *FileTools) GetTools() ([]agentcore.Tool, error) {
	if ft == nil || ft.securedFs == nil {
		return nil, fmt.Errorf("文件工具未初始化")
	}

	var tools []agentcore.Tool

	fileReadTool, err := agentcore.InferTool("file_read", fileReadDesc, ft.FileRead)
	if err != nil {
		return nil, err
	}
	tools = append(tools, fileReadTool)

	fileWriteTool, err := agentcore.InferTool("file_write", fileWriteDesc, ft.FileWrite)
	if err != nil {
		return nil, err
	}
	tools = append(tools, fileWriteTool)

	fileAppendTool, err := agentcore.InferTool("file_append", fileAppendDesc, ft.FileAppend)
	if err != nil {
		return nil, err
	}
	tools = append(tools, fileAppendTool)

	fileEditTool, err := agentcore.InferTool("file_edit", fileEditDesc, ft.FileEdit)
	if err != nil {
		return nil, err
	}
	tools = append(tools, fileEditTool)

	grepTool, err := agentcore.InferTool("grep", grepDesc, ft.Grep)
	if err != nil {
		return nil, err
	}
	tools = append(tools, grepTool)

	fileListTool, err := agentcore.InferTool("file_list", fileListDesc, ft.FileList)
	if err != nil {
		return nil, err
	}
	tools = append(tools, fileListTool)

	globTool, err := agentcore.InferTool("glob", globDesc, ft.Glob)
	if err != nil {
		return nil, err
	}
	tools = append(tools, globTool)

	filePatchTool, err := agentcore.InferTool("file_patch", filePatchDesc, ft.FilePatch)
	if err != nil {
		return nil, err
	}
	tools = append(tools, filePatchTool)

	return tools, nil
}
