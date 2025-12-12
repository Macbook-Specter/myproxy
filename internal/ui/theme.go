package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// MonochromeTheme 实现 Fyne 主题接口，提供黑白两套主题（Dark/Light）。
// 该主题使用简化的配色方案，但保持良好的可读性和对比度。
type MonochromeTheme struct {
	variant fyne.ThemeVariant
}

// NewMonochromeTheme 创建黑白主题实例。
// 参数：
//   - variant: 主题变体，支持 fyne.ThemeVariantDark（黑色）或 fyne.ThemeVariantLight（白色）
//
// 返回：主题实例
func NewMonochromeTheme(variant fyne.ThemeVariant) fyne.Theme {
	return &MonochromeTheme{variant: variant}
}

// Color 返回自定义颜色，未覆盖的颜色使用默认主题
func (t *MonochromeTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// 以传入 variant 优先，其次使用主题自身 variant
	if variant == fyne.ThemeVariant(0) {
		variant = t.variant
	}

	switch variant {
	case theme.VariantDark:
		switch name {
		case theme.ColorNameBackground, theme.ColorNameInputBackground:
			return color.NRGBA{R: 18, G: 18, B: 18, A: 255} // 近黑背景
		case theme.ColorNameForeground:
			return color.NRGBA{R: 220, G: 220, B: 220, A: 255} // 淡白色前景，适合日志区域显示
		case theme.ColorNameButton, theme.ColorNamePrimary:
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255} // 纯白色按钮和主要元素
		case theme.ColorNameFocus, theme.ColorNameHover:
			return color.NRGBA{R: 255, G: 255, B: 255, A: 64} // 半透明高亮
		}
	case theme.VariantLight:
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255} // 白色背景
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 245, G: 245, B: 245, A: 255} // 浅灰输入背景，避免过强对比
		case theme.ColorNameForeground:
			return color.NRGBA{R: 30, G: 30, B: 30, A: 255} // 深色文字
		case theme.ColorNameButton:
			return color.NRGBA{R: 235, G: 235, B: 235, A: 255} // 浅灰按钮背景，保持白色主题
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 40, G: 40, B: 40, A: 255} // 轻微强调色，避免纯黑导致按钮过暗
		case theme.ColorNameFocus, theme.ColorNameHover:
			return color.NRGBA{R: 0, G: 0, B: 0, A: 48} // 半透明高亮，降低一点强度
		}
	}

	// 其他颜色使用默认主题
	return theme.DefaultTheme().Color(name, variant)
}

// Icon 使用默认主题图标
func (t *MonochromeTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Font 使用默认字体，保持兼容
func (t *MonochromeTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Size 使用默认尺寸
func (t *MonochromeTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
