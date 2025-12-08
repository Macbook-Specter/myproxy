package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"myproxy.com/p/internal/database"
)

// SubscriptionPanel 订阅管理面板
// 使用双向绑定自动更新标签显示
type SubscriptionPanel struct {
	appState      *AppState
	tagContainer  fyne.CanvasObject // 标签容器（使用 HBox 以便动态更新）
	headerArea    fyne.CanvasObject // 头部区域（包含标签容器）
	subscriptions []*database.Subscription
}

// NewSubscriptionPanel 创建订阅管理面板
func NewSubscriptionPanel(appState *AppState) *SubscriptionPanel {
	sp := &SubscriptionPanel{
		appState: appState,
	}

	// 创建标签容器（水平布局）
	sp.tagContainer = container.NewHBox()

	// 加载订阅列表
	sp.refreshSubscriptionList()

	// 监听绑定数据变化，自动更新标签显示
	appState.SubscriptionLabelsBinding.AddListener(binding.NewDataListener(func() {
		sp.updateTagsFromBinding()
	}))

	return sp
}

// Build 构建订阅面板 UI
func (sp *SubscriptionPanel) Build() fyne.CanvasObject {
	// 从绑定数据初始化标签显示
	sp.updateTagsFromBinding()

	// 加号按钮
	addBtn := widget.NewButton("+", sp.onAddSubscription)

	// 订阅管理标题和标签组
	sp.headerArea = container.NewHBox(
		widget.NewLabel("订阅管理"),
		sp.tagContainer,
		addBtn,
	)

	return container.NewVBox(
		sp.headerArea,
		widget.NewSeparator(),
	)
}

// updateTagsFromBinding 从绑定数据更新标签显示（使用双向绑定）
func (sp *SubscriptionPanel) updateTagsFromBinding() {
	// 从绑定数据获取标签列表
	labels, err := sp.appState.SubscriptionLabelsBinding.Get()
	if err != nil {
		// 如果获取失败，从数据库重新加载
		sp.refreshSubscriptionList()
		sp.appState.UpdateSubscriptionLabels()
		return
	}

	// 获取所有订阅（用于创建按钮的回调）
	sp.refreshSubscriptionList()

	// 创建新的标签按钮列表
	var tagButtons []fyne.CanvasObject

	// 为每个标签创建按钮
	for _, label := range labels {
		// 找到对应的订阅
		var sub *database.Subscription
		for _, s := range sp.subscriptions {
			if s.Label == label {
				sub = s
				break
			}
		}

		if sub != nil {
			// 创建标签按钮，点击时弹出编辑对话框
			tagBtn := widget.NewButton(label, func(s *database.Subscription) func() {
				return func() {
					sp.onEditSubscription(s)
				}
			}(sub))
			tagButtons = append(tagButtons, tagBtn)
		}
	}

	// 重新创建容器
	sp.tagContainer = container.NewHBox(tagButtons...)

	// 刷新 headerArea（如果已创建）
	// 注意：由于 Fyne 容器的不可变性，我们需要在主窗口级别刷新
	// 这里我们只是更新 tagContainer，主窗口会在需要时刷新
}

// refreshTags 刷新标签显示（保留用于兼容性，现在使用绑定）
func (sp *SubscriptionPanel) refreshTags() {
	// 更新绑定数据，UI 会自动更新
	if sp.appState != nil {
		sp.appState.UpdateSubscriptionLabels()
	}
}

// onEditSubscription 编辑订阅（弹出对话框）
func (sp *SubscriptionPanel) onEditSubscription(sub *database.Subscription) {
	// 创建对话框内容
	urlEntry := widget.NewEntry()
	urlEntry.SetText(sub.URL)
	urlEntry.SetPlaceHolder("请输入订阅URL（必填）")

	labelEntry := widget.NewEntry()
	labelEntry.SetText(sub.Label)
	labelEntry.SetPlaceHolder("请输入标签（必填）")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "订阅URL", Widget: urlEntry, HintText: "必填项"},
			{Text: "标签", Widget: labelEntry, HintText: "必填项"},
		},
	}

	// 创建对话框
	dialog.ShowForm("编辑订阅", "确定", "取消", form.Items, func(confirmed bool) {
		if !confirmed {
			return
		}

		url := urlEntry.Text
		label := labelEntry.Text

		// 验证必填项
		if url == "" {
			dialog.ShowError(fmt.Errorf("订阅URL不能为空"), sp.appState.Window)
			return
		}
		if label == "" {
			dialog.ShowError(fmt.Errorf("标签不能为空"), sp.appState.Window)
			return
		}

		// 如果URL改变，更新订阅
		if url != sub.URL {
			// 更新订阅
			err := sp.appState.SubscriptionManager.UpdateSubscription(url, label)
			if err != nil {
				dialog.ShowError(fmt.Errorf("订阅更新失败: %w", err), sp.appState.Window)
				return
			}
		} else if label != sub.Label {
			// 只更新标签
			_, err := database.AddOrUpdateSubscription(url, label)
			if err != nil {
				dialog.ShowError(fmt.Errorf("标签更新失败: %w", err), sp.appState.Window)
				return
			}
		}

		// 刷新订阅列表
		sp.refreshSubscriptionList()
		// 更新绑定数据，UI 会自动更新
		sp.appState.UpdateSubscriptionLabels()

		sp.appState.Window.SetTitle("订阅已更新")
	}, sp.appState.Window)
}

// onAddSubscription 添加订阅（弹出对话框）
func (sp *SubscriptionPanel) onAddSubscription() {
	// 创建对话框内容
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("请输入订阅URL（必填）")

	labelEntry := widget.NewEntry()
	labelEntry.SetPlaceHolder("请输入标签（必填）")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "订阅URL", Widget: urlEntry, HintText: "必填项"},
			{Text: "标签", Widget: labelEntry, HintText: "必填项"},
		},
	}

	// 创建对话框
	dialog.ShowForm("添加订阅", "确定", "取消", form.Items, func(confirmed bool) {
		if !confirmed {
			return
		}

		url := urlEntry.Text
		label := labelEntry.Text

		// 验证必填项
		if url == "" {
			dialog.ShowError(fmt.Errorf("订阅URL不能为空"), sp.appState.Window)
			return
		}
		if label == "" {
			dialog.ShowError(fmt.Errorf("标签不能为空"), sp.appState.Window)
			return
		}

		// 获取订阅
		servers, err := sp.appState.SubscriptionManager.FetchSubscription(url, label)
		if err != nil {
			dialog.ShowError(fmt.Errorf("订阅获取失败: %w", err), sp.appState.Window)
			return
		}

		// 刷新订阅列表
		sp.refreshSubscriptionList()
		// 更新绑定数据，UI 会自动更新
		sp.appState.UpdateSubscriptionLabels()

		sp.appState.Window.SetTitle(fmt.Sprintf("订阅添加成功，共 %d 条服务器", len(servers)))
	}, sp.appState.Window)
}

// refreshSubscriptionList 刷新订阅列表
func (sp *SubscriptionPanel) refreshSubscriptionList() {
	subscriptions, err := database.GetAllSubscriptions()
	if err != nil {
		// 如果数据库未初始化，使用空列表
		sp.subscriptions = []*database.Subscription{}
	} else {
		sp.subscriptions = subscriptions
	}
	// 注意：不再在这里刷新标签，而是通过绑定自动更新
}
