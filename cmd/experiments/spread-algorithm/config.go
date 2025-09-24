package main

import (
    "fmt"
    "os"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Version  string   `yaml:"version"`
    Defaults Defaults `yaml:"defaults"`
    Spreads  []Spread `yaml:"spreads"`
}

type Defaults struct {
    Paper    Paper    `yaml:"paper"`
    Margins  Margins  `yaml:"margins"`
    Spread   SpreadOpt `yaml:"spread"`
    Crop     Crop     `yaml:"crop"`
    Scale    Scale    `yaml:"scale"`
    Position Position `yaml:"position"`
    Export   Export   `yaml:"export"`
}

type Spread struct {
    Name     string     `yaml:"name"`
    Paper    *Paper     `yaml:"paper"`
    Margins  *Margins   `yaml:"margins"`
    Spread   *SpreadOpt `yaml:"spread"`
    Crop     *Crop      `yaml:"crop"`
    Scale    *Scale     `yaml:"scale"`
    Position *Position  `yaml:"position"`
    Export   *Export    `yaml:"export"`
}

type Paper struct {
    WidthIn     float64 `yaml:"width_in"`
    HeightIn    float64 `yaml:"height_in"`
    Orientation string  `yaml:"orientation"` // portrait | landscape
    DPI         float64 `yaml:"dpi"`
}

type Margins struct {
    AllIn   *float64 `yaml:"all_in"`
    TopIn   float64  `yaml:"top_in"`
    RightIn float64  `yaml:"right_in"`
    BottomIn float64 `yaml:"bottom_in"`
    LeftIn  float64  `yaml:"left_in"`
}

type SpreadOpt struct {
    IsSpread bool    `yaml:"is_spread"`
    GutterIn float64 `yaml:"gutter_in"`
}

type Crop struct {
    Ratio  string `yaml:"ratio"`   // "original" | "w:h" | ""
    ToFill bool   `yaml:"to_fill"` // true => cover
}

type Scale struct {
    UserScale float64 `yaml:"user_scale"`
}

type Position struct {
    X     float64 `yaml:"x"`
    Y     float64 `yaml:"y"`
    Units string  `yaml:"units"` // px | normalized
}

type Export struct {
    Format           string `yaml:"format"` // png | jpg | pdf (pdf not implemented)
    Quality          int    `yaml:"quality"`
    Background       string `yaml:"background"` // "transparent" | #hex | css name (limited)
    OutDir           string `yaml:"out_dir"`
    FilenameTemplate string `yaml:"filename_template"`
    PNGLevel         string `yaml:"png_level"`
    Scaler           string `yaml:"scaler"`          // fast | quality
    ParallelEncode   bool   `yaml:"parallel_encode"`
}

type ResolvedSpread struct {
    Name     string
    Paper    Paper
    Margins  Margins
    Spread   SpreadOpt
    Crop     Crop
    Scale    Scale
    Position Position
    Export   Export
}

func LoadConfig(path string) (*Config, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var c Config
    if err := yaml.Unmarshal(b, &c); err != nil {
        return nil, err
    }
    return &c, nil
}

func (c *Config) MergeWithDefaults(sp Spread) ResolvedSpread {
    rs := ResolvedSpread{
        Name:     sp.Name,
        Paper:    c.Defaults.Paper,
        Margins:  c.Defaults.Margins,
        Spread:   c.Defaults.Spread,
        Crop:     c.Defaults.Crop,
        Scale:    c.Defaults.Scale,
        Position: c.Defaults.Position,
        Export:   c.Defaults.Export,
    }

    if sp.Paper != nil {
        if sp.Paper.WidthIn > 0 { rs.Paper.WidthIn = sp.Paper.WidthIn }
        if sp.Paper.HeightIn > 0 { rs.Paper.HeightIn = sp.Paper.HeightIn }
        if sp.Paper.Orientation != "" { rs.Paper.Orientation = sp.Paper.Orientation }
        if sp.Paper.DPI > 0 { rs.Paper.DPI = sp.Paper.DPI }
    }
    if sp.Margins != nil {
        if sp.Margins.AllIn != nil {
            v := *sp.Margins.AllIn
            rs.Margins.TopIn, rs.Margins.RightIn, rs.Margins.BottomIn, rs.Margins.LeftIn = v, v, v, v
        } else {
            // Override each side (note: zeros are valid and will override)
            rs.Margins.TopIn = sp.Margins.TopIn
            rs.Margins.RightIn = sp.Margins.RightIn
            rs.Margins.BottomIn = sp.Margins.BottomIn
            rs.Margins.LeftIn = sp.Margins.LeftIn
        }
    } else if rs.Margins.AllIn != nil {
        v := *rs.Margins.AllIn
        rs.Margins.TopIn, rs.Margins.RightIn, rs.Margins.BottomIn, rs.Margins.LeftIn = v, v, v, v
    }
    // Ensure AllIn is nil in resolved to avoid confusion
    rs.Margins.AllIn = nil

    if sp.Spread != nil {
        rs.Spread.IsSpread = sp.Spread.IsSpread
        rs.Spread.GutterIn = sp.Spread.GutterIn
    }
    if sp.Crop != nil {
        if sp.Crop.Ratio != "" { rs.Crop.Ratio = sp.Crop.Ratio }
        rs.Crop.ToFill = sp.Crop.ToFill
    }
    if sp.Scale != nil {
        rs.Scale.UserScale = sp.Scale.UserScale
    }
    if sp.Position != nil {
        rs.Position.X = sp.Position.X
        rs.Position.Y = sp.Position.Y
        if sp.Position.Units != "" { rs.Position.Units = sp.Position.Units }
    }
    if sp.Export != nil {
        if sp.Export.Format != "" { rs.Export.Format = sp.Export.Format }
        if sp.Export.Quality != 0 { rs.Export.Quality = sp.Export.Quality }
        if sp.Export.Background != "" { rs.Export.Background = sp.Export.Background }
        if sp.Export.OutDir != "" { rs.Export.OutDir = sp.Export.OutDir }
        if sp.Export.FilenameTemplate != "" { rs.Export.FilenameTemplate = sp.Export.FilenameTemplate }
        if sp.Export.PNGLevel != "" { rs.Export.PNGLevel = sp.Export.PNGLevel }
        if sp.Export.Scaler != "" { rs.Export.Scaler = sp.Export.Scaler }
        rs.Export.ParallelEncode = sp.Export.ParallelEncode
    }

    return rs
}

// ToInputs converts the resolved spread into algorithm inputs
func (rs ResolvedSpread) ToInputs(srcW, srcH float64) Inputs {
    mt, mr, mb, ml := rs.Margins.TopIn, rs.Margins.RightIn, rs.Margins.BottomIn, rs.Margins.LeftIn
    // crop ratio parse
    var cropRatio *CropRatio
    switch rs.Crop.Ratio {
    case "", "original":
        cropRatio = nil
    default:
        var w, h float64
        if _, err := fmt.Sscanf(rs.Crop.Ratio, "%f:%f", &w, &h); err == nil && w > 0 && h > 0 {
            cropRatio = &CropRatio{W: w, H: h}
        }
    }

    return Inputs{
        SrcW: srcW,
        SrcH: srcH,
        PaperWIn: rs.Paper.WidthIn,
        PaperHIn: rs.Paper.HeightIn,
        Orientation: rs.Paper.Orientation,
        MarginTopIn: mt,
        MarginRightIn: mr,
        MarginBottomIn: mb,
        MarginLeftIn: ml,
        DPI: rs.Paper.DPI,
        IsSpread: rs.Spread.IsSpread,
        GutterIn: rs.Spread.GutterIn,
        CropRatio: cropRatio,
        CropToFill: rs.Crop.ToFill,
        UserScale: rs.Scale.UserScale,
        ImagePosition: Point{X: rs.Position.X, Y: rs.Position.Y},
        PositionUnits: rs.Position.Units,
    }
}


