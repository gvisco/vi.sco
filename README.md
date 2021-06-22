# vi.sco
Vito Scognamiglio a.k.a. Vi.Sco - A Telegram bot written in Golang

## See also
[Golang bindings for the Telegram Bot API
](https://pkg.go.dev/github.com/go-telegram-bot-api/telegram-bot-api@v4.6.4+incompatible?utm_source=gopls#section-readme)

[Effective Go](https://golang.org/doc/effective_go)

graph LR
    Waiting -->|/list help| Help
    Waiting -->|/list all| ListAll
    Waiting -->|/list view <name>| ViewList
    Waiting -->|/list new <name>| NewList
    Waiting -->|/list del <name>| DeleteListConfirm
    Waiting -->|/list edit <name>| EditList
    Waiting -->|/list <unrecognized>| Help

    Help -->|<nil>| Waiting
    
    ListAll -->|<nil>| Waiting
    
    ViewList -->|<nil>| Waiting
    
    NewList -->|<nil>| NewInput

    NewInput -->|/end| NewDone
    NewInput -->|*| NewInput
    
    NewDone -->|<nil>| Waiting

    DeleteListConfirm --> |<nil>| DeleteListConfirmInput
    
    DeleteListConfirmInput -->|no| Waiting
    DeleteListConfirmInput -->|yes| DeleteListDone
    DeleteListConfirmInput -->|*| DeleteListConfirmInput
    
    DeleteListDone -->|<nil>| Waiting
    
    EditList -->|<nil>| EditInput

    EditInput -->|/end| EditDone
    EditInput -->|/help| EditHelp
    EditInput -->|/view| EditView
    EditInput -->|/append <item>| EditAppend
    EditInput -->|/rm <position>| EditRemove
    EditInput -->|/add <position> <item>| EditAdd
    EditInput -->|/mv <from> <to>| EditMove
    EditInput -->|/edit <position> <item>| EditEdit
    EditInput -->|*| EditInvalid

    EditHelp -->|nil| EditInput

    EditView -->|<nil>| EditInput

    EditAppend -->|<nil>| EditInput

    EditRemove -->|<nil>| EditInput

    EditAdd -->|<nil>| EditInput

    EditMove -->|<nil>| EditInput

    EditEdit -->|<nil>| EditInput

    EditInvalid -->|<nil>| EditInput

    EditDone -->|<nil>| Waiting