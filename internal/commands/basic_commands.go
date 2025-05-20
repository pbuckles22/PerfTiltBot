package commands

// RegisterBasicCommands registers all basic queue management commands
func RegisterBasicCommands(cm *CommandManager) {
	cm.RegisterCommand(&Command{
		Name:        "help",
		Description: "Show the list of available commands",
		Handler:     handleHelp,
	})

	cm.RegisterCommand(&Command{
		Name:        "ping",
		Description: "Check if the bot is alive",
		Handler:     handlePing,
	})

	cm.RegisterCommand(&Command{
		Name:        "sq",
		Description: "Start the queue system",
		Handler:     handleStartQueue,
	})

	cm.RegisterCommand(&Command{
		Name:        "endqueue",
		Description: "End the queue system",
		Handler:     handleEndQueue,
	})

	cm.RegisterCommand(&Command{
		Name:        "clearqueue",
		Description: "Clear all users from the queue",
		Handler:     handleClearQueue,
	})

	cm.RegisterCommand(&Command{
		Name:        "queue",
		Description: "Show the current queue",
		Handler:     handleQueue,
	})

	cm.RegisterCommand(&Command{
		Name:        "join",
		Aliases:     []string{"j"},
		Description: "Join the queue",
		Handler:     handleJoin,
	})

	cm.RegisterCommand(&Command{
		Name:        "leave",
		Aliases:     []string{"l"},
		Description: "Leave the queue",
		Handler:     handleLeave,
	})

	cm.RegisterCommand(&Command{
		Name:        "list",
		Aliases:     []string{"q"},
		Description: "Show the current queue",
		Handler:     handleList,
	})

	cm.RegisterCommand(&Command{
		Name:        "position",
		Aliases:     []string{"pos"},
		Description: "Show your position in the queue",
		Handler:     handlePosition,
	})

	cm.RegisterCommand(&Command{
		Name:        "pop",
		Aliases:     []string{"p"},
		Description: "Pop users from the queue",
		Handler:     handlePop,
	})

	cm.RegisterCommand(&Command{
		Name:        "move",
		Aliases:     []string{"m"},
		Description: "Move a user in the queue",
		Handler:     handleMove,
	})

	cm.RegisterCommand(&Command{
		Name:        "remove",
		Aliases:     []string{"r"},
		Description: "Remove a user from the queue",
		Handler:     handleRemove,
	})

	cm.RegisterCommand(&Command{
		Name:        "clear",
		Aliases:     []string{"c"},
		Description: "Clear the queue",
		Handler:     handleClear,
	})

	cm.RegisterCommand(&Command{
		Name:        "enable",
		Aliases:     []string{"e"},
		Description: "Enable the queue system",
		Handler:     handleEnable,
	})

	cm.RegisterCommand(&Command{
		Name:        "disable",
		Aliases:     []string{"d"},
		Description: "Disable the queue system",
		Handler:     handleDisable,
	})

	cm.RegisterCommand(&Command{
		Name:        "pause",
		Aliases:     []string{"pa"},
		Description: "Pause the queue system",
		Handler:     handlePause,
	})

	cm.RegisterCommand(&Command{
		Name:        "unpause",
		Aliases:     []string{"up"},
		Description: "Unpause the queue system",
		Handler:     handleUnpause,
	})

	cm.RegisterCommand(&Command{
		Name:        "save",
		Aliases:     []string{"s"},
		Description: "Save the queue state",
		Handler:     handleSaveState,
	})

	cm.RegisterCommand(&Command{
		Name:        "load",
		Aliases:     []string{"ld"},
		Description: "Load the queue state",
		Handler:     handleLoadState,
	})

	cm.RegisterCommand(&Command{
		Name:        "kill",
		Aliases:     []string{"k"},
		Description: "Shutdown the bot",
		Handler:     handleKill,
	})

	cm.RegisterCommand(&Command{
		Name:        "restart",
		Aliases:     []string{"rs"},
		Description: "Restart the bot",
		Handler:     handleRestart,
	})
}

// SaveState saves the current queue state
func (cm *CommandManager) SaveState() error {
	return cm.queue.SaveState()
}

// LoadState loads the queue state
func (cm *CommandManager) LoadState() error {
	return cm.queue.LoadState()
}
