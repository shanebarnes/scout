package main

import (
    "errors"

    gc "github.com/rthornton128/goncurses"
)

type MenuNcurses struct {
    impl MenuImpl
    win *gc.Window
    menu *gc.Menu
    items []*gc.MenuItem
    currentIndex int
}

func (m *MenuNcurses) New(h, w, x, y int) error {
    m.Delete()

    m.impl.winHeight = h
    m.impl.winWidth = w
    m.impl.winX = x
    m.impl.winY = y

    m.items = make([]*gc.MenuItem, len(m.impl.children))

    //for i := 0; i < len(m.impl.children); i++ {
    //    j := (m.currentIndex + i) % len(m.impl.children)
    //    child := *(m.impl.children[j])
    //    m.items[i], _ = gc.NewItem(child.GetImpl().title, "")
    //}
    for i, child := range m.impl.children {
        m.items[i], _ = gc.NewItem((*child).GetImpl().title, "")
    }

    m.menu, _ = gc.NewMenu(m.items)
    m.win, _ = gc.NewWindow(h, w, y, x)
    m.win.Keypad(true)
    m.menu.SetWindow(m.win)
    dwin := m.win.Derived(h - 4, w - 2, 3, 1)
    m.menu.SubWindow(dwin)
    m.menu.Format(h - 4, 1)
    m.menu.Mark(" * ")

//    if m.currentIndex >= 0 && m.currentIndex < len(m.items) {
//        m.menu.Current(m.items[m.currentIndex])
//    }

    _, maxX := m.win.MaxYX()
    m.win.Box(0, 0)
    m.win.ColorOn(gc.C_CYAN)
    m.win.MovePrint(1, (maxX / 2 ) - (len(m.impl.title) / 2), m.impl.title)
    m.win.ColorOff(gc.C_CYAN)
    m.win.MoveAddChar(2, 1, gc.ACS_LTEE)
    m.win.HLine(2, 2, gc.ACS_HLINE, maxX - 3)
    m.win.MoveAddChar(2, maxX - 2, gc.ACS_RTEE)

    m.menu.Post()
    m.win.Refresh()
    go MenuThread(m)

    return nil
}

func (m *MenuNcurses) Delete() error {
    if m.menu != nil {
        m.menu.UnPost()

        if m.items != nil {
            for i, _ := range m.items {
                m.items[i].Free()
            }

            m.items = nil
        }

        m.menu.Free()
        m.menu = nil
    }

    return nil
}

func (m *MenuNcurses) SetTitle(title string) {
    m.impl.title = title
}

func (m *MenuNcurses) SetParent(parent *Menu) error {
    var err error = nil

    if parent == nil {
        err = errors.New("Invalid parent menu")
    } else {
        m.impl.parent = parent
    }

    return err
}

func (m *MenuNcurses) AddChild(child *Menu) error {
    var err error = nil

    if child == nil {
        err = errors.New("Invalid child menu")
    } else {
        m.impl.children = append(m.impl.children, child)
    }

    return err
}

func (m *MenuNcurses) GetImpl() *MenuImpl {
    return &m.impl
}

func MenuThread(m *MenuNcurses) {
    for {
        //gc.Update()
m.win.Refresh()
        char := m.win.GetChar()

        switch char {
            case gc.KEY_EXIT, gc.KEY_LEFT:
                if m.impl.parent != nil {
                    m.currentIndex = m.menu.Current(nil).Index()
                    m.Delete()
                    (*m.impl.parent).New(m.impl.winHeight, m.impl.winWidth, m.impl.winX, m.impl.winY)
                    return
                } else {
                    m.menu.Driver(gc.KEY_LEFT)
                }
            case gc.KEY_RETURN, gc.KEY_RIGHT:
                child := *m.impl.children[m.currentIndex]
                if len(child.GetImpl().children) > 0 {
                    m.currentIndex = m.menu.Current(nil).Index()
                    m.Delete()
                    (*m.impl.children[m.currentIndex]).New(m.impl.winHeight, m.impl.winWidth, m.impl.winX, m.impl.winY)
                    return
                } else {
                    m.menu.Driver(gc.KEY_RIGHT)
                }
            case gc.KEY_DOWN:
                m.currentIndex = m.menu.Current(nil).Index()
                m.menu.Driver(gc.REQ_DOWN)
                if len(m.impl.children) > 1 && m.currentIndex + 1 == len(m.impl.children) {
                    m.menu.Current(m.items[0])
                }
                m.currentIndex = m.menu.Current(nil).Index()
            case gc.KEY_PAGEDOWN:
                m.menu.Driver(gc.REQ_PAGE_DOWN)
            case gc.KEY_PAGEUP:
                m.menu.Driver(gc.REQ_PAGE_UP)
            case gc.KEY_UP:
                m.currentIndex = m.menu.Current(nil).Index()
                m.menu.Driver(gc.REQ_UP)
                //if m.currentIndex == 0 && len(m.impl.children) > 1 {
                //    m.menu.Current(m.items[len(m.impl.children) - 1])
                //}
                //m.currentIndex = m.menu.Current(nil).Index()
        }
    }
}
