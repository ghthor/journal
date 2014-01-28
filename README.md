This program is intended to help me write and enjoy writing my personal journal.
The `journal` stores each entry in a git repository.

## Persistence from entry to entry

After using the journal for a few days I found myself needing something that persists from entry to entry.
I'm calling this an `Idea`.
The following is the syntax in a go template.

```
## [{{.Status}}] {{.Title}}
{{.Body}}
```

An Idea's status can be

| Status |     |
| :----: | --- |
| active | Carried over from entry to entry |
| inactive | Ignored when creating a new entry |
| completed | Not used yet, is treated exactly like [inactive]. Will be used in the future. |

To view how Idea's look in use view an entry in the `log/`.
For instance, the commit where I added this documentation I transfered all the tasks from this readme into Idea's stored in the log.
Most of them were marked as inactive because they are not the focus of work on the project at this time.
But now they are stored in the log and I'll be able to resurrect [inactive] ideas in the future using `journal`.
