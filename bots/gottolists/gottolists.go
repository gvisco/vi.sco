package gottolists

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gvisco/vi.sco/gotto"
)

const noEvent string = "<nil evt>"

const helpString string = `Available commands:
/list all -- Print the names of all the available lists
/list view <name> -- Print the content of the a list
/list new <name> -- Create a new list with given name
/list del <name> -- Delete a list
/list edit <name> -- Edit the content of a list
/list help -- Print this help message
`

var reListView *regexp.Regexp = regexp.MustCompile(`/list view ([^ ]+)$`)
var reNewList *regexp.Regexp = regexp.MustCompile(`/list new ([^ ]+)$`)
var reDelList *regexp.Regexp = regexp.MustCompile(`/list del ([^ ]+)$`)
var reEditList *regexp.Regexp = regexp.MustCompile(`/list edit ([^ ]+)$`)
var reUnrecognizedList *regexp.Regexp = regexp.MustCompile(`/list(.+)$`)

type state int

const (
	waiting state = iota
	help
	listAll
	viewList
	newList
	newInput
	newDone
	deleteListConfirm
	deleteListDone
	editList
	editInput
	editView
	editAppend
	editRemove
	editAdd
	editMove
	editEdit
	editInvalid
	editDone
)

func (s state) String() string {
	switch s {
	case waiting:
		return "Waiting"
	case help:
		return "Hep"
	case listAll:
		return "ListAll"
	case viewList:
		return "ViewList"
	case newList:
		return "NewList"
	case newInput:
		return "NewInput"
	case newDone:
		return "NewDone"
	case deleteListConfirm:
		return "DeleteListConfirm"
	case deleteListDone:
		return "DeleteListDone"
	case editList:
		return "EditList"
	case editInput:
		return "EditInput"
	case editView:
		return "EditView"
	case editAppend:
		return "EditAppend"
	case editRemove:
		return "EditRemove"
	case editAdd:
		return "EditAdd"
	case editMove:
		return "EditMove"
	case editEdit:
		return "EditEdit"
	case editInvalid:
		return "EditInvalid"
	case editDone:
		return "EditDone"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}

type StateMachine struct {
	current state
	nodes   map[state]Node
}

type Node struct {
	activate func(*ListBot, string) string
	edges    []Edge
}

type Edge struct {
	matcher func(string) bool
	dest    state
}

func initStateMachine() *StateMachine {
	return &StateMachine{
		current: waiting,
		nodes: map[state]Node{
			waiting: {
				activate: func(bot *ListBot, msg string) string { return "" },
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == "/list help" },
						dest:    help,
					},
					{
						matcher: func(s string) bool { return s == "/list all" },
						dest:    listAll,
					},
					{
						matcher: func(s string) bool { return reListView.MatchString(s) },
						dest:    viewList,
					},
					{
						matcher: func(s string) bool { return reNewList.MatchString(s) },
						dest:    newList,
					},
					{
						matcher: func(s string) bool { return reDelList.MatchString(s) },
						dest:    deleteListConfirm,
					},
					{
						matcher: func(s string) bool { return reEditList.MatchString(s) },
						dest:    editList,
					},
					{
						matcher: func(s string) bool { return reUnrecognizedList.MatchString(s) },
						dest:    help,
					},
				},
			},
			help: {
				activate: func(lb *ListBot, s string) string { return helpString },
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    waiting,
					},
				},
			},
			listAll: {
				activate: func(lb *ListBot, s string) string {
					result := "Your lists:"
					for _, l := range lb.lists {
						result = fmt.Sprintf("%s\n- %s", result, l.name)
					}
					return result
				},
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    waiting,
					},
				},
			},
			viewList: {
				activate: func(lb *ListBot, s string) string {
					lname := reListView.FindStringSubmatch(s)[1]
					l, ok := lb.lists[lname]
					if !ok {
						return fmt.Sprintf("Invalid list name: %s", lname)
					}
					result := fmt.Sprintf("--- %s ---", lname)
					for idx, val := range l.items {
						result = fmt.Sprintf("%s\n[%d] %s", result, idx, val)
					}
					return result
				},
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    waiting,
					},
				},
			},
			newList: {
				activate: func(lb *ListBot, s string) string {
					lname := reNewList.FindStringSubmatch(s)[1]
					_, ok := lb.lists[lname]
					if ok {
						lb.state.current = waiting
						return fmt.Sprintf("A list with name '%s' already exists", lname)
					}
					list := &List{
						name:     lname,
						filePath: lb.workspace + "/" + lname + ".list",
						items:    []string{},
					}
					err := list.saveToFile()
					if err != nil {
						lb.state.current = waiting
						log.Printf("[ERROR ListBot Cannot save list to file] Workspace {%s} ListName {%s} Error {%s} ", lb.workspace, lname, err)
						return fmt.Sprintf("Cannot save list '%s'. An error occurred", lname)
					}
					lb.lists[lname] = list
					lb.currentList = list
					return fmt.Sprintf("I'm listening. Add new items to list '%s'.\nWrite `/end` to complete", lname)
				},
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    newInput,
					},
				},
			},
			newInput: {
				activate: func(lb *ListBot, s string) string {
					lb.currentList.addItem(s)
					lname := lb.currentList.name
					err := lb.currentList.saveToFile()
					if err != nil {
						lb.state.current = waiting
						log.Printf("[ERROR ListBot Cannot save list to file] Workspace {%s} ListName {%s} Error {%s} ", lb.workspace, lname, err)
						return fmt.Sprintf("Cannot save list '%s'. An error occurred", lname)
					}
					return ""
				},
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == "/end" },
						dest:    newDone,
					},
					{
						matcher: func(s string) bool { return true },
						dest:    newInput,
					},
				},
			},
			newDone: {
				activate: func(lb *ListBot, s string) string {
					return fmt.Sprintf("New list '%s' created with %d items", lb.currentList.name, len(lb.currentList.items))
				},
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    waiting,
					},
				},
			},
			deleteListConfirm: {
				activate: func(lb *ListBot, s string) string {
					lname := reDelList.FindStringSubmatch(s)[1]
					l, ok := lb.lists[lname]
					if !ok {
						return fmt.Sprintf("Invalid list name: %s", lname)
					}
					lb.currentList = l
					return fmt.Sprintf("Are you sure you want to delete list '%s'?\nPlease reply 'yes' or 'no'", lb.currentList.name)
				},
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == "no" },
						dest:    waiting,
					},
					{
						matcher: func(s string) bool { return s == "yes" },
						dest:    deleteListDone,
					},
					{
						// TODO: fix it! It does not work as the input is going to replace the current list
						matcher: func(s string) bool { return true },
						dest:    deleteListConfirm,
					},
				},
			},
			deleteListDone: {
				activate: func(lb *ListBot, s string) string {
					toBeDeleted := lb.currentList
					err := os.Remove(toBeDeleted.filePath)
					if err != nil {
						lb.state.current = waiting
						log.Printf("[ERROR ListBot Cannot delete list file] Workspace {%s} ListName {%s} Error {%s} ", lb.workspace, toBeDeleted.name, err)
						return fmt.Sprintf("Cannot delete list '%s'. An error occurred", toBeDeleted.name)
					}
					lb.currentList = nil
					delete(lb.lists, toBeDeleted.name)
					return fmt.Sprintf("List '%s' succesfully deleted", toBeDeleted.name)
				},
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    waiting,
					},
				},
			},
			editList: {
				activate: func(lb *ListBot, s string) string { return "TODO: edit list" },
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    editInput,
					},
				},
			},
			editInput: {
				activate: func(lb *ListBot, s string) string { return "TODO: edit input" },
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == "/end" },
						dest:    editDone,
					},
					{
						matcher: func(s string) bool { return s == "/view" },
						dest:    editView,
					},
					{
						matcher: func(s string) bool { return match("/append (.+)", s) },
						dest:    editAppend,
					},
					{
						matcher: func(s string) bool { return match(`/rm (\d+)`, s) },
						dest:    editRemove,
					},
					{
						matcher: func(s string) bool { return match(`/add (\d+) (.+)`, s) },
						dest:    editAdd,
					},
					{
						matcher: func(s string) bool { return match(`/mv (\d+) (\d+)`, s) },
						dest:    editMove,
					},
					{
						matcher: func(s string) bool { return match(`/edit (\d+) (.+)`, s) },
						dest:    editEdit,
					},
					{
						matcher: func(s string) bool { return true },
						dest:    editInvalid,
					},
				},
			},
			editView: {
				activate: func(lb *ListBot, s string) string { return "TODO: edit view" },
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    editInput,
					},
				},
			},
			editAppend: {
				activate: func(lb *ListBot, s string) string { return "TODO: edit append" },
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    editInput,
					},
				},
			},
			editRemove: {
				activate: func(lb *ListBot, s string) string { return "TODO: edit remove" },
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    editInput,
					},
				},
			},
			editAdd: {
				activate: func(lb *ListBot, s string) string { return "TODO: edit add" },
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    editInput,
					},
				},
			},
			editMove: {
				activate: func(lb *ListBot, s string) string { return "TODO: edit move" },
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    editInput,
					},
				},
			},
			editEdit: {
				activate: func(lb *ListBot, s string) string { return "TODO: edit edit" },
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    editInput,
					},
				},
			},
			editInvalid: {
				activate: func(lb *ListBot, s string) string { return "TODO: edit invalid" },
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    editInput,
					},
				},
			},
			editDone: {
				activate: func(lb *ListBot, s string) string { return "TODO: edit done" },
				edges: []Edge{
					{
						matcher: func(s string) bool { return s == noEvent },
						dest:    waiting,
					},
				},
			},
		},
	}
}

func match(re string, s string) bool {
	b, err := regexp.MatchString(re, s)
	if err != nil {
		log.Printf("[ERROR ListBot cannot apply matching] Regexp {%s} Input {%s} Error {%s}", re, s, err)
		return false
	}

	return b
}

type ListBotFactory struct {
}

type ListBot struct {
	workspace   string
	lists       map[string]*List
	state       *StateMachine
	currentList *List
}

type List struct {
	name     string
	filePath string
	items    []string
}

func NewFactory() *ListBotFactory {
	return &ListBotFactory{}
}

func (*ListBotFactory) CreateBot(workspace string) (gotto.GottoBot, error) {
	log.Printf("[Create new ListBot] Workspace {%s}", workspace)
	sm := initStateMachine()
	// read existing lists
	files, err := ioutil.ReadDir(workspace)
	if err != nil {
		log.Printf("[ERROR ListBot cannot read files in workspace] Workspace {%s} Error {%s}", workspace, err)
		return nil, err
	}

	lists := make(map[string]*List)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".list") {
			fileName := workspace + "/" + file.Name()
			name := strings.TrimSuffix(file.Name(), filepath.Ext(fileName))
			list := &List{name: name, filePath: fileName}
			err = list.loadFromFile()
			if err != nil {
				log.Printf("[ERROR ListBot cannot read list from file] File {%s} Error {%s}", fileName, err)
				continue
			}
			lists[name] = list
		}
	}

	log.Printf("[ListBot created] Workspace {%s} Lists {%d}", workspace, len(lists))
	return &ListBot{workspace: workspace, lists: lists, state: sm, currentList: nil}, nil
}

func (bot *ListBot) OnUpdate(userId string, userName string, message string) string {
	curr := bot.state.nodes[bot.state.current]
	for _, e := range curr.edges {
		if e.matcher(message) {
			log.Printf("[ListBot changing state] From {%+v} To {%+v}", bot.state.current, e.dest)
			bot.state.current = e.dest
			node := bot.state.nodes[e.dest]
			reply := node.activate(bot, message)
			// try the <nil> move
			for _, e2 := range node.edges {
				if e2.matcher(noEvent) {
					log.Printf("[ListBot changing state] From {%+v} To {%+v}", bot.state.current, e.dest)
					bot.state.current = e2.dest
				}
			}
			return reply
		}
	}
	return ""
}

func (list *List) addItem(s string) {
	list.items = append(list.items, s)
}

func (list *List) loadFromFile() error {
	file, err := os.Open(list.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	list.items = lines

	return scanner.Err()
}

func (list *List) saveToFile() error {
	file, err := os.Create(list.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range list.items {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
