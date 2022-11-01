package minigo

import (
	"fmt"
)

type Service struct {
	homePage Page
	subPages map[uint]Page
}

func NewService(homePage Page) *Service {
	return &Service{
		homePage: homePage,
		subPages: make(map[uint]Page),
	}
}

func (s *Service) RegisterSubPage(id uint, subPage Page) error {
	if _, ok := s.subPages[id]; ok {
		return fmt.Errorf("page id %d already registered", id)
	}

	s.subPages[id] = subPage
	return nil
}

func (s *Service) NewConnection(driver Driver) {
	s.homePage.NewSession(driver)
}
