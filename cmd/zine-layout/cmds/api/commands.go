package api

import (
	"github.com/spf13/cobra"
)

// AddAllAPICommands adds all API commands to the provided root command
func AddAllAPICommands(rootCmd *cobra.Command) error {
	// Projects commands
	if err := AddProjectsListCommand(rootCmd); err != nil {
		return err
	}
	if err := AddProjectsGetCommand(rootCmd); err != nil {
		return err
	}
	if err := AddProjectsCreateCommand(rootCmd); err != nil {
		return err
	}
	if err := AddProjectsDeleteCommand(rootCmd); err != nil {
		return err
	}

	// Images commands
	if err := AddImagesListCommand(rootCmd); err != nil {
		return err
	}

	// Presets commands
	if err := AddPresetsListCommand(rootCmd); err != nil {
		return err
	}

	return nil
}

// AddProjectsListCommand adds the projects-list command
func AddProjectsListCommand(cmd *cobra.Command) error {
	projectsListCmd, err := NewProjectsListCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := buildAPICommand(projectsListCmd)
	if err != nil {
		return err
	}

	cmd.AddCommand(cobraCmd)
	return nil
}

// AddProjectsGetCommand adds the projects-get command
func AddProjectsGetCommand(cmd *cobra.Command) error {
	projectsGetCmd, err := NewProjectsGetCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := buildAPICommand(projectsGetCmd)
	if err != nil {
		return err
	}

	cmd.AddCommand(cobraCmd)
	return nil
}

// AddProjectsCreateCommand adds the projects-create command
func AddProjectsCreateCommand(cmd *cobra.Command) error {
	projectsCreateCmd, err := NewProjectsCreateCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := buildAPICommand(projectsCreateCmd)
	if err != nil {
		return err
	}

	cmd.AddCommand(cobraCmd)
	return nil
}

// AddProjectsDeleteCommand adds the projects-delete command
func AddProjectsDeleteCommand(cmd *cobra.Command) error {
	projectsDeleteCmd, err := NewProjectsDeleteCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := buildAPICommand(projectsDeleteCmd)
	if err != nil {
		return err
	}

	cmd.AddCommand(cobraCmd)
	return nil
}

// AddImagesListCommand adds the images-list command
func AddImagesListCommand(cmd *cobra.Command) error {
	imagesListCmd, err := NewImagesListCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := buildAPICommand(imagesListCmd)
	if err != nil {
		return err
	}

	cmd.AddCommand(cobraCmd)
	return nil
}

// AddPresetsListCommand adds the presets-list command
func AddPresetsListCommand(cmd *cobra.Command) error {
	presetsListCmd, err := NewPresetsListCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := buildAPICommand(presetsListCmd)
	if err != nil {
		return err
	}

	cmd.AddCommand(cobraCmd)
	return nil
}
