# journal

A tool to write a journal that is stored on the file system in a git repository

## Overview

## Usage

### Installing

The following `go get` will install a single combined binary.

    $ go get github.com/ghthor/journal

```
someone@somewhere $ journal help
journal is a wrapper around git for creating a project/personal log.

Usage:

	journal command [command arguments]

The commands are:
    init       initialize a new journal directory
    fix        upgrade the storage format
    new        create, edit, and save an entry to a journal

```

Each individual command is also available as a stand alone
binary.

    $ go get github.com/ghthor/journal/exec/journal-init
    $ go get github.com/ghthor/journal/exec/journal-new

### Using journal

#### Initialization

To begin writing in the journal you first need to initialize
a directory as a journal.

    $ journal init path/to/directory

You can also initialize a journal directory that is within an existing
git repository. `journal` will use the existing git repository to
store all the changes that are made as you add entries and ideas.
You can view the [log/](https://github.com/ghthor/journal/tree/master/log)
for the journal project to see it working in action.

After initializing a journal it will have the following directory tree.

```
journal/
├── entry
└── idea
    ├── active
    └── nextid

2 directories, 2 files
```

The `entry/` directory stores each journal entry in a filename
based on the date the entry was open. The `idea/` directory
stores a document type that is persistent from entry to entry.

#### Create an Entry in the journal

With an initialized directory you can now begin adding entries
to the journal.

    $ journal new path/to/directory

This will open a new entry template in a text editor. The only
implemented editor is vim. If you are interested in adding support
for another editor please open a ticket or pull request.

You must save the file before exiting the editor. `journal` uses a
real file on the file system and if it is not saved it will commit
an empty template.

When you exit the editor, `journal` will commit the entry and any
ideas(the persistent document type) to the repository. Review the
git commit log and the contents of `entry/` and `idea/` directories
to view how your entry is stored and committed.

### Using Ideas

TODO

## Contributing

1. Fork it
2. `go get github.com/ghthor/journal`
3. Add your fork as a remote to `$GOPATH/src/github.com/ghthor/journal`
4. Create your feature branch (`git checkout -b my-new-feature`)
5. Commit your changes (`git commit -am 'Add some feature'`)
6. Push to the branch (`git push fork-remote my-new-feature`)
7. Create new Pull Request

## License

journal is released under the MIT license. See [LICENSE](https://github.com/ghthor/journal/blob/master/LICENSE)
