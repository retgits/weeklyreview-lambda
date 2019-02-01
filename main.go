package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/adlio/trello"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	boardID   string
	listID    string
	labelmap  map[string]string
	statusmap map[string]string
	cardsmap  map[string][]*trello.Card
)

const (
	version         = "0.1.0"
	boardName       = "Main"
	listName        = "Done"
	plannedStatus   = "Planned / Done"
	notDoneStatus   = "Planned / Not Done"
	unplannedStatus = "Unplanned / Done"
	unknownStatus   = "Other Tasks"
	wentWellTitle   = "What went well"
	beBetterTitle   = "What didn't go well"
	needHelpTitle   = "Where do I need help"

	tpl = `Weekly Review
Planned Done
These are items / tasks that were planned last week, and executed exactly as planned. Good job! Lets reward somebody and sing some praises!
{{ range .PlannedDone }}- {{ .Name }} {{ if .Labels }}({{ getLabels .IDLabels }}){{ end }}
{{ end }}
Planned / Not Done
These are items / tasks that were planned for this week, but were not executed. The questions to asks are: What happened? Why? Who was responsible? What are the next steps?
{{ range .PlannedNotDone }}- {{ .Name }} {{ if .Labels }}({{ getLabels .IDLabels }}){{ end }}
{{ end }}
Unplanned / Done
These are items / tasks that were not planned, but were finished. This is not necessarily a good thing! Why was this worked on if it wasn’t planned? Who’s responsible?
{{ range .UnplannedDone }}- {{ .Name }} {{ if .Labels }}({{ getLabels .IDLabels }}){{ end }}
{{ end }}
Other tasks
{{ range .Unknown }}- {{ .Name }} {{ if .Labels }}({{ getLabels .IDLabels }}){{ end }}
{{ end }}
What went well
{{ .WentWell }}

What didn't go well
{{ .BeBetter }}

Where do I need help
{{ .NeedHelp }}
`
)

type config struct {
	AppKey     string `required:"true"`
	AppToken   string `required:"true"`
	MemberName string `required:"true"`
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msgf("Running version %s of weekly-review", version)

	envConfig := config{}

	log.Info().Msg("Reading environment variables")
	err := envconfig.Process("trello", &envConfig)
	if err != nil {
		log.Fatal().Msgf("Fatal error while reading configuration: %s", err.Error())
	}

	log.Info().Msg("Connecting to Trello")
	client := trello.NewClient(envConfig.AppKey, envConfig.AppToken)

	log.Info().Msg("Getting member details")
	member, err := client.GetMember(envConfig.MemberName, trello.Defaults())
	if err != nil {
		log.Fatal().Msgf("Error while getting Trello member details: %s", err.Error())
	}

	log.Info().Msg("Getting Trello boards")
	boards, err := member.GetBoards(trello.Defaults())
	if err != nil {
		log.Fatal().Msgf("Error while getting Trello boards: %s", err.Error())
	}

	log.Info().Msgf("Finding Trello board %s", boardName)
	for idx := range boards {
		log.Debug().Msgf("Current board: %s", boards[idx].Name)
		if boards[idx].Name == boardName {
			boardID = boards[idx].ID
			break
		}
	}

	log.Info().Msg("Getting Trello board")
	board, err := client.GetBoard(boardID, trello.Defaults())
	if err != nil {
		log.Fatal().Msgf("Error while getting Trello board: %s", err.Error())
	}

	log.Info().Msg("Getting Trello lists")
	lists, err := board.GetLists(trello.Defaults())
	if err != nil {
		log.Fatal().Msgf("Error while getting Trello lists: %s", err.Error())
	}

	log.Info().Msgf("Finding Trello list %s", listName)
	for idx := range lists {
		log.Debug().Msgf("Current list: %s", lists[idx].Name)
		if lists[idx].Name == listName {
			listID = lists[idx].ID
			break
		}
	}

	log.Info().Msg("Getting Trello list")
	list, err := client.GetList(listID, trello.Defaults())
	if err != nil {
		log.Fatal().Msgf("Error while getting Trello list: %s", err.Error())
	}

	log.Info().Msg("Getting Trello labels")
	labels, err := board.GetLabels(trello.Defaults())
	if err != nil {
		log.Error().Msg(err.Error())
	}

	labelmap = make(map[string]string)

	for idx := range labels {
		label := labels[idx]
		labelmap[label.ID] = label.Name
	}

	log.Info().Msg("Getting Trello cards")
	cards, err := list.GetCards(trello.Defaults())
	if err != nil {
		log.Fatal().Msgf("Error while getting Trello cards: %s", err.Error())
	}

	log.Info().Msg("Putting Trello cards in categories")
	plannedCardsArray := make([]*trello.Card, 0)
	notdoneCardsArray := make([]*trello.Card, 0)
	unplannedCardsArray := make([]*trello.Card, 0)
	unknownCardsArray := make([]*trello.Card, 0)

	for idx := range cards {
		category := getCategory(cards[idx])
		switch category {
		case plannedStatus:
			plannedCardsArray = append(plannedCardsArray, cards[idx])
		case notDoneStatus:
			notdoneCardsArray = append(notdoneCardsArray, cards[idx])
		case unplannedStatus:
			unplannedCardsArray = append(unplannedCardsArray, cards[idx])
		default:
			unknownCardsArray = append(unknownCardsArray, cards[idx])
		}
	}

	log.Info().Msg("Preparing data")
	var wentWell string
	var beBetter string
	var needHelp string
	needHelp, unknownCardsArray = getCard(needHelpTitle, unknownCardsArray)
	beBetter, unknownCardsArray = getCard(beBetterTitle, unknownCardsArray)
	wentWell, unknownCardsArray = getCard(wentWellTitle, unknownCardsArray)

	data := struct {
		PlannedDone    []*trello.Card
		PlannedNotDone []*trello.Card
		UnplannedDone  []*trello.Card
		Unknown        []*trello.Card
		WentWell       string
		BeBetter       string
		NeedHelp       string
	}{
		PlannedDone:    plannedCardsArray,
		PlannedNotDone: notdoneCardsArray,
		UnplannedDone:  unplannedCardsArray,
		Unknown:        unknownCardsArray,
		WentWell:       wentWell,
		BeBetter:       beBetter,
		NeedHelp:       needHelp,
	}

	var buff bytes.Buffer

	funcMap := template.FuncMap{
		"getLabels": getTemplateLabels,
	}

	log.Info().Msg("Parsing template")
	parsedTpl, err := template.New("tpl").Funcs(funcMap).Parse(tpl)
	if err != nil {
		log.Fatal().Msgf("Error while preparing text template: %s", err.Error())
	}

	if err := parsedTpl.Execute(&buff, data); err != nil {
		log.Fatal().Msgf("Error while parsing text template: %s", err.Error())
	}

	log.Info().Msgf("\n\n%s\n\n", buff.String())
}

func getCategory(card *trello.Card) string {
	labels := getLabels(card.IDLabels)
	if strings.Contains(labels, plannedStatus) {
		return plannedStatus
	} else if strings.Contains(labels, notDoneStatus) {
		return notDoneStatus
	} else if strings.Contains(labels, unplannedStatus) {
		return unplannedStatus
	}
	return unknownStatus
}

func getTemplateLabels(labelID []string) string {
	labels := getLabels(labelID)
	labels = strings.Replace(labels, plannedStatus, "", -1)
	labels = strings.Replace(labels, notDoneStatus, "", -1)
	labels = strings.Replace(labels, unplannedStatus, "", -1)
	labels = strings.TrimSpace(labels)
	labels = strings.Replace(labels, " ", ", ", -1)
	return labels
}

func getLabels(labelID []string) string {
	var labels string
	for idx := range labelID {
		labels = fmt.Sprintf("%s %s", labels, labelmap[labelID[idx]])
	}
	return labels
}

func getCard(title string, cards []*trello.Card) (string, []*trello.Card) {
	for idx := range cards {
		if cards[idx].Name == title {
			return cards[idx].Desc, remove(cards, idx)
		}
	}
	return "", cards
}

func remove(s []*trello.Card, i int) []*trello.Card {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
