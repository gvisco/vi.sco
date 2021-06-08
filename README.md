# vi.sco
Vito Scognamiglio a.k.a. Vi.Sco - A Telegram bot written in Golang

## See also
[Golang bindings for the Telegram Bot API
](https://pkg.go.dev/github.com/go-telegram-bot-api/telegram-bot-api@v4.6.4+incompatible?utm_source=gopls#section-readme)

[Effective Go](https://golang.org/doc/effective_go)

graph TD
    Waiting -->|/list help| Help
    Waiting -->|/list all| ListAll
    Waiting -->|/list view <name>| ViewList
    Waiting -->|/list new <name>| NewList
    Waiting -->|/list del <name>| ConfirmDel
    Waiting -->|/list edit <name>| EditList
    Waiting -->|/list <unrecognized>| Help

    Help -->|<nil>| Waiting
    
    ListAll -->|<nil>| Waiting
    
    ViewList -->|<nil>| Waiting
    
    NewList -->|<nil>| AddItem
    
    AddItem -->|*| AddItem
    AddItem -->|/end| Waiting
    
    ConfirmDel -->|no| Waiting
    ConfirmDel -->|yes| DelList

    DelList -->|<nil>| Waiting
    
    EditList -->|<nil>| EditItem
    
    EditItem -->|/append <item>| EditItem
    EditItem -->|/rm <position>| EditItem
    EditItem -->|/add <position> <item>| EditItem
    EditItem -->|/mv <from> <to>| EditItem
    EditItem -->|/edit <position> <item>| EditItem
    EditItem -->|/end| Waiting