package api

import (
	imagelayouttemplates "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/api/image_layout_templates"
	imagesequences "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/api/image_sequences"
	laidoutimages "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/api/laid_out_images"
	layoutsequences "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/api/layout_sequences"
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
	seqCmd, err := imagesequences.NewCommand()
	if err != nil {
		return err
	}
	rootCmd.AddCommand(seqCmd)

	tplCmd, err := imagelayouttemplates.NewCommand()
	if err != nil {
		return err
	}
	rootCmd.AddCommand(tplCmd)

	imgCmd, err := laidoutimages.NewCommand()
	if err != nil {
		return err
	}
	rootCmd.AddCommand(imgCmd)

	seqLayoutCmd, err := layoutsequences.NewCommand()
	if err != nil {
		return err
	}
	rootCmd.AddCommand(seqLayoutCmd)

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

// legacy commands removed in new workflow
