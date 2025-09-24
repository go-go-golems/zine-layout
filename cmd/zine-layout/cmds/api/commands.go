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
	if err := AddImagesUploadCommand(rootCmd); err != nil {
		return err
	}
	if err := AddImagesUploadDirCommand(rootCmd); err != nil {
		return err
	}
	if err := AddImagesSyncCommand(rootCmd); err != nil {
		return err
	}

	// Presets commands
	if err := AddPresetsListCommand(rootCmd); err != nil {
		return err
	}

	// Pages commands
	if err := AddPagesListCommand(rootCmd); err != nil {
		return err
	}
	if err := AddPagesGetCommand(rootCmd); err != nil {
		return err
	}
	if err := AddPagesPutCommand(rootCmd); err != nil {
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

// AddImagesUploadCommand adds the images-upload command
func AddImagesUploadCommand(cmd *cobra.Command) error {
	imagesUploadCmd, err := NewImagesUploadCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := buildAPICommand(imagesUploadCmd)
	if err != nil {
		return err
	}

	cmd.AddCommand(cobraCmd)
	return nil
}

// AddImagesUploadDirCommand adds the images-upload-dir command
func AddImagesUploadDirCommand(cmd *cobra.Command) error {
	imagesUploadDirCmd, err := NewImagesUploadDirCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := buildAPICommand(imagesUploadDirCmd)
	if err != nil {
		return err
	}

	cmd.AddCommand(cobraCmd)
	return nil
}

// AddImagesSyncCommand adds the images-sync command
func AddImagesSyncCommand(cmd *cobra.Command) error {
	imagesSyncCmd, err := NewImagesSyncCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := buildAPICommand(imagesSyncCmd)
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

// AddPagesListCommand adds the pages-list command
func AddPagesListCommand(cmd *cobra.Command) error {
	pagesListCmd, err := NewPagesListCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := buildAPICommand(pagesListCmd)
	if err != nil {
		return err
	}

	cmd.AddCommand(cobraCmd)
	return nil
}

// AddPagesGetCommand adds the pages-get command
func AddPagesGetCommand(cmd *cobra.Command) error {
	pagesGetCmd, err := NewPagesGetCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := buildAPICommand(pagesGetCmd)
	if err != nil {
		return err
	}

	cmd.AddCommand(cobraCmd)
	return nil
}

// AddPagesPutCommand adds the pages-put command
func AddPagesPutCommand(cmd *cobra.Command) error {
	pagesPutCmd, err := NewPagesPutCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := buildAPICommand(pagesPutCmd)
	if err != nil {
		return err
	}

	cmd.AddCommand(cobraCmd)
	return nil
}
