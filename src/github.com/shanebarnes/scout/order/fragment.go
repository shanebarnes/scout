package order

import (
    "errors"
)

type FragmentImpl struct {
    // mutex
}

type Fragment interface {
    New() error
    Parse()
    GetImpl() *FragmentImpl
}
