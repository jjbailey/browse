# browse

A simple, unconventional file browser designed for efficient file navigation and viewing.

## Description

**browse** is a minimalist file browser that focuses on essential features while maintaining a user-friendly interface. It's ideal for developers and system administrators who need a lightweight, keyboard-driven alternative to traditional file viewers.

The browser supports multi-file browsing by accepting a list of files (or globs expanded by your shell), ensuring each path is resolved and validated before browsing. Input via pipe is supported and handled as if it were a temporary file. Invalid files, such as directories or non-existent files, are safely skipped with feedback.

The browser enables you to temporarily leave your current file list to browse a new set of files; when you finish browsing the new list, **browse** automatically returns you to where you left off in the previous list. This "recursive file list" feature lets you explore deeply and flexibly without losing your place.

## ✨ Features

### 🧭 Navigation

- Forward and reverse paging
- Continuous scrolling (forward and reverse)
- Horizontal scrolling
- Line jumping
- Page marking

### 🔍 Search & Filter

- Forward and reverse regex searches
- Case-sensitive/case-insensitive search toggle
- Pattern highlighting
- Search pattern history

### 🛠️ Additional Features

- File completion
- Shell escape with command completion
- `tail -f` functionality
- Line numbers
- Session saving
- Shell command history
- Change working directory
- Built-in help screen

## 📖 Usage

### Command Line Options

```bash
browse [OPTIONS] [FILE] [FILE...]
```

| Option                | Function                                          |
| --------------------- | ------------------------------------------------- |
| `-f`, `--follow`      | Follow file changes (like `tail -f`)              |
| `-i`, `--ignore-case` | Search ignores case                               |
| `-n`, `--numbers`     | Start with line numbers turned on                 |
| `-p`, `--pattern`     | Initial search pattern                            |
| `-t`, `--title`       | Page title (default is filename, blank for stdin) |
| `-v`, `--version`     | Print browse version number                       |
| `-?`, `--help`        | Print browse command line options                 |

### Keyboard Shortcuts

#### Navigation

| Key                       | Function                                           |
| ------------------------- | -------------------------------------------------- |
| `f`, `Page Down`, `Space` | Page down toward EOF                               |
| `b`, `Page Up`            | Page up toward SOF                                 |
| `Ctrl+F`, `Ctrl+D`, `z`   | Scroll half page down toward EOF                   |
| `Ctrl+B`, `Ctrl+U`, `Z`   | Scroll half page up toward SOF                     |
| `+`, `Right`, `Enter`     | Scroll one line toward EOF                         |
| `-`, `Left`               | Scroll one line toward SOF                         |
| `d`, `Down`               | Toggle continuous scroll toward EOF, follow at EOF |
| `u`, `Up`                 | Toggle continuous scroll toward SOF, stop at SOF   |
| `>`, `Tab`                | Scroll 4 characters right                          |
| `<`, `Backspace`, `Del`   | Scroll 4 characters left                           |
| `^`                       | Scroll to column 1                                 |
| `$`                       | Scroll to end of line                              |
| `0`, `Home`               | Jump to start of file, column 1                    |
| `G`                       | Jump to end of file                                |
| `e`, `End`                | Jump to EOF, follow at EOF                         |
| `t`                       | Jump to EOF, tail at EOF                           |

#### Search

| Key | Function                                                            |
| --- | ------------------------------------------------------------------- |
| `/` | Regex search forward (empty pattern repeats or changes direction)   |
| `?` | Regex search reverse (empty pattern repeats or changes direction)   |
| `n` | Repeat search in current direction                                  |
| `N` | Repeat search in opposite direction                                 |
| `i` | Toggle case-sensitive/insensitive search                            |
| `p` | Print current search pattern                                        |
| `P` | Clear search pattern                                                |
| `&` | Run `grep -nP` on current file for search pattern in a new session  |

#### Miscellaneous

| Key           | Function                                             |
| ------------- | ---------------------------------------------------- |
| `#`           | Toggle line numbers on/off                           |
| `%`, `Ctrl+G` | Show page position                                   |
| `!`           | Run a bash command (expands `!`, `%`, `&`, `~`)      |
| `F`           | Run `fmt -s` on the current file in a new session    |
| `B`           | Browse another file (expands `%`, `~`, shell glob)   |
| `a`           | Print filenames in the browse list                   |
| `c`           | Print current working directory                      |
| `C`           | Change working directory                             |
| `q`           | Quit, save session, next file in list                |
| `Q`           | Quit without saving session, next file in list       |
| `x`           | Exit list, save session                              |
| `X`           | Exit list, don't save session                        |

### Symbol Expansions

Special symbols expanded in commands:

| Symbol | Expands To             |
| ------ | ---------------------- |
| `!`    | Last bash command      |
| `%`    | Current file name      |
| `&`    | Current search pattern |
| `~`    | Home directory         |

## ⚙️ Configuration

browse stores configuration and history in `~/.browse/`

**Session File:** `~/.browse/browserc`

Saves current session state, including:

1. File name
2. First line on page
3. Search pattern
4. Marks
5. Page title

## 🕑 Histories

**browse** maintains persistent histories to streamline your workflow and make repeated tasks faster and easier. If you're familiar with Linux tools like bash, these history features will feel instantly familiar:

- **Bash Command History:** Every time you run a shell command from within browse (using the `!` key), it's remembered. You can quickly recall, edit, and rerun previous commands—just like using the up-arrow in bash.
- **Directory History:** Whenever you change your working directory within browse, it's recorded. You can quickly cycle through or return to recently used directories, helping you navigate large projects or complex file trees more efficiently.
- **File History:** Files browsed are saved (as full pathnames) for easy access in the current or future sessions.
- **Search Pattern History:** Search patterns entered for regex or text searches are saved. This feature lets you easily repeat searches or revisit common queries without retyping them.

**History Files:**

- `~/.browse/browse_dirs` - Directory history
- `~/.browse/browse_files` - File browsing history
- `~/.browse/browse_search` - Search pattern history
- `~/.browse/browse_shell` - Shell command history

## ⚠️ Limitations

- Xterm specific
- Logical lines are truncated to screen width
- May be US-centric
- Can be confused by non-printable characters
- Tabs are converted to spaces
- Terminal title changes due to go-prompt dependency

## 📄 License

MIT License - see LICENSE file for details
