package main

type MenuImpl struct {
    title       string
    description string
    parent      *Menu
    children    []*Menu
}

type Menu interface {
    New() error
    Delete() error
    SetTitle(title string)
    SetParent(parent *Menu) error
    AddChild(child *Menu) error
    GetImpl() *MenuImpl
}
