package laidoutpagescmd

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/settings"
    "github.com/go-go-golems/glazed/pkg/types"
    workflowshared "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/workflow/shared"
    "github.com/go-go-golems/zine-layout/pkg/services"
    "github.com/spf13/cobra"
)

type laidOutPagesRenderCommand struct{
    *cmds.CommandDescription
}

type laidOutPagesRenderSettings struct{
    DataRoot string `glazed.parameter:"data-root"`
    PageID   string `glazed.parameter:"page-id"`
}

func (c *laidOutPagesRenderCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    s := &laidOutPagesRenderSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    if s.PageID == "" {
        return fmt.Errorf("--page-id is required")
    }
    repos, db, err := workflowshared.OpenRepositories(s.DataRoot)
    if err != nil { return err }
    defer db.Close()
    ps := services.NewPagesService(repos)
    ps.SetDataRoot(s.DataRoot)
    page, err := ps.RenderPage(s.PageID)
    if err != nil { return err }
    var meta struct{
        Variants map[string]string `json:"variants"`
    }
    if page.ResultJSON != nil {
        _ = json.Unmarshal([]byte(*page.ResultJSON), &meta)
    }
    row := types.NewRow(
        types.MRP("entity", "laid_out_page_render"),
        types.MRP("page_id", page.ID),
        types.MRP("project_id", page.ProjectID),
        types.MRP("thumbnail", meta.Variants["thumbnail"]),
        types.MRP("full", meta.Variants["full"]),
        types.MRP("combined", meta.Variants["combined"]),
        types.MRP("left", meta.Variants["left"]),
        types.MRP("right", meta.Variants["right"]),
    )
    return gp.AddRow(ctx, row)
}

func newLaidOutPagesRenderCommand()(*cobra.Command, error){
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil { return nil, err }
    cmd := &laidOutPagesRenderCommand{
        CommandDescription: cmds.NewCommandDescription(
            "render",
            cmds.WithShort("Render a laid-out page and print variant paths"),
            cmds.WithFlags(
                parameters.NewParameterDefinition("data-root", parameters.ParameterTypeString, parameters.WithDefault("./data"), parameters.WithHelp("Path to data directory")),
                parameters.NewParameterDefinition("page-id", parameters.ParameterTypeString, parameters.WithRequired(true), parameters.WithHelp("Laid-out page ID")),
            ),
            cmds.WithLayersList(glazedLayer),
        ),
    }
    return workflowshared.BuildCommand(cmd)
}

var _ cmds.GlazeCommand = &laidOutPagesRenderCommand{}


