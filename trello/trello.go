package trello

import (
	"fmt"

	"github.com/adlio/trello"
)

var (
	tLabels map[string]string
)

// Service is a struct that represents a Trello connection
type Service struct {
	AppKey     string
	AppToken   string
	MemberName string
	client     *trello.Client
}

// GetCardInput is the input for the GetCard method
type GetCardInput struct {
	Board      string
	List       string
	WithLabels bool
}

// GetCardOutput is the output for the GetCard method
type GetCardOutput struct {
	Cards []*Card
}

// Card is the individual Trello card as a response
type Card struct {
	Name        string
	Description string
	Labels      []string
}

// New creates a new instance of the service
func New() *Service {
	return &Service{}
}

// WithKey allows the AppKey to be set
func (s *Service) WithKey(key string) *Service {
	s.AppKey = key
	return s
}

// WithToken allows the AppKey to be set
func (s *Service) WithToken(token string) *Service {
	s.AppToken = token
	return s
}

// WithMemberName allows the AppKey to be set
func (s *Service) WithMemberName(membername string) *Service {
	s.MemberName = membername
	return s
}

func (s *Service) GetCards(gci *GetCardInput) (*GetCardOutput, error) {
	if s.client == nil {
		s.client = trello.NewClient(s.AppKey, s.AppToken)
	}

	member, err := s.client.GetMember(s.MemberName, trello.Defaults())
	if err != nil {
		return nil, fmt.Errorf("Error while getting Trello member details: %w", err)
	}

	boards, err := member.GetBoards(trello.Defaults())
	if err != nil {
		return nil, fmt.Errorf("Error while getting Trello boards: %w", err)
	}

	var boardID string

	for idx := range boards {
		if boards[idx].Name == gci.Board {
			boardID = boards[idx].ID
			break
		}
	}

	if len(boardID) == 0 {
		return nil, fmt.Errorf("No board '%s' found", gci.Board)
	}

	board, err := s.client.GetBoard(boardID, trello.Defaults())
	if err != nil {
		return nil, fmt.Errorf("Error while getting Trello board: %w", err)
	}

	lists, err := board.GetLists(trello.Defaults())
	if err != nil {
		return nil, fmt.Errorf("Error while getting Trello lists: %w", err)
	}

	var listID string

	for idx := range lists {
		if lists[idx].Name == gci.List {
			listID = lists[idx].ID
			break
		}
	}

	if len(listID) == 0 {
		return nil, fmt.Errorf("No list '%s' found", gci.List)
	}

	list, err := s.client.GetList(listID, trello.Defaults())
	if err != nil {
		return nil, fmt.Errorf("Error while getting Trello list: %w", err)
	}

	cards, err := list.GetCards(trello.Defaults())
	if err != nil {
		return nil, fmt.Errorf("Error while getting Trello cards: %w", err)
	}

	if gci.WithLabels {
		labels, err := board.GetLabels(trello.Defaults())
		if err != nil {
			return nil, fmt.Errorf("Error while getting Trello labels: %w", err)
		}

		tLabels = make(map[string]string)

		for idx := range labels {
			label := labels[idx]
			tLabels[label.ID] = label.Name
		}
	}

	c := make([]*Card, len(cards))

	for idx := range cards {
		card := &Card{
			Name:        cards[idx].Name,
			Description: cards[idx].Desc,
		}

		if gci.WithLabels {
			card.Labels = cardLabels(cards[idx].IDLabels)
		}

		c[idx] = card
	}

	return &GetCardOutput{
		Cards: c,
	}, nil
}

// cardLabels matches the IDs of the labels with the actual name and returns an
// array of names
func cardLabels(labelID []string) []string {
	labels := make([]string, len(labelID))
	for idx := range labelID {
		labels[idx] = tLabels[labelID[idx]]
	}
	return labels
}
