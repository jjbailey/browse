# browse

A simple, unconventional file browser designed for efficient file navigation and viewing.

## Description

**browse** is a minimalist file browser that focuses on essential features while maintaining a user-friendly interface. It's ideal for developers and system administrators who need a lightweight, keyboard-driven alternative to traditional file viewers.

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

- Shell escape with command completion
- `tail -f` functionality
- Line numbers
- Session saving
- Shell command history
- Built-in help screen

## 📖 Usage

### Command Line Options

```bash
browse [OPTIONS] [FILE] [FILE...]
```

| Option | Function |
|--------|----------|
| `-f`, `--follow` | Follow file changes (like `tail -f`) |
| `-i`, `--ignore-case` | Search ignores case |
| `-n`, `--numbers` | Start with line numbers turned on |
| `-p`, `--pattern` | Initial search pattern |
| `-t`, `--title` | Page title (default is filename, blank for stdin) |
| `-v`, `--version` | Print browse version number |
| `-?`, `--help` | Print browse command line options |

### Keyboard Shortcuts

#### Navigation

| Key | Function |
|-----|----------|
| `f`, `Page Down`, `Space` | Page down toward EOF |
| `b`, `Page Up` | Page up toward SOF |
| `Ctrl+F`, `Ctrl+D`, `z` | Scroll half page down toward EOF |
| `Ctrl+B`, `Ctrl+U`, `Z` | Scroll half page up toward SOF |
| `+`, `Right`, `Enter` | Scroll one line toward EOF |
| `-`, `Left` | Scroll one line toward SOF |
| `d`, `Down` | Toggle continuous scroll toward EOF, follow at EOF |
| `u`, `Up` | Toggle continuous scroll toward SOF, stop at SOF |
| `>`, `Tab` | Scroll 4 characters right |
| `<`, `Backspace`, `Del` | Scroll 4 characters left |
| `^` | Scroll to column 1 |
| `$` | Scroll to end of line |
| `0`, `Home` | Jump to start of file, column 1 |
| `G` | Jump to end of file |
| `e`, `End` | Jump to EOF, follow at EOF |
| `t` | Jump to EOF, tail at EOF |

#### Search

| Key | Function |
|-----|----------|
| `/` | Regex search forward (empty pattern repeats or changes direction) |
| `?` | Regex search reverse (empty pattern repeats or changes direction) |
| `n` | Repeat search in current direction |
| `N` | Repeat search in opposite direction |
| `i` | Toggle case-sensitive/insensitive search |
| `p` | Print current search pattern |
| `P` | Clear search pattern |
| `&` | Run `grep -nP` on current file for search pattern |

#### Miscellaneous

| Key | Function |
|-----|----------|
| `#` | Toggle line numbers on/off |
| `%`, `Ctrl+G` | Show page position |
| `!` | Run a bash command (expands `!`, `%`, `&`, `~`) |
| `B` | Browse another file (expands `%`, `~`) |
| `q` | Quit, save session, next file in list |
| `Q` | Quit without saving session, next file in list |
| `x` | Exit list, save session |
| `X` | Exit list, don't save session |

### Symbol Expansions

Special symbols are expanded in commands:

| Symbol | Expands To |
|--------|------------|
| `!` | Last bash command |
| `%` | Current file name |
| `&` | Current search pattern |
| `~` | Home directory |

## ⚙️ Configuration

browse stores configuration and history in `~/.browse/`:

**Session File:** `~/.browse/browserc`

Saves current session state including:

1. File name
2. First line on page
3. Search pattern
4. Marks
5. Page title

**History Files:**

- `~/.browse/browse_files` - File browsing history
- `~/.browse/browse_shell` - Shell command history
- `~/.browse/browse_search` - Search pattern history

## ⚠️ Limitations

- Xterm specific
- Logical lines are truncated to screen width
- May be US-centric
- Can be confused by non-printable characters
- Tabs are converted to spaces
- Terminal title changes due to go-prompt dependency

## 📄 License

MIT License - see LICENSE file for details
