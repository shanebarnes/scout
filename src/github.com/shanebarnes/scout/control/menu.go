package control

type MenuImpl struct {
    title       string
    description string
    parent      *Menu
    children    []*Menu
    winHeight, winWidth, winX, winY int
}

type Menu interface {
    New(h, w, x, y int) error
    Delete() error
    SetTitle(title string)
    SetParent(parent *Menu) error
    AddChild(child *Menu) error
    GetImpl() *MenuImpl
}
