# browse

A simple, unconventional file browser designed for efficient file navigation and viewing.

## Description

browse is a minimalist file browser that focuses on essential features while maintaining a user-friendly interface. It's ideal for developers and system administrators who need a lightweight alternative to traditional file viewers.

## Features

### Navigation

- Forward and reverse paging
- Continuous scrolling (forward and reverse)
- Horizontal scrolling
- Line jumping
- Page marking

### Search & Filter

- Forward and reverse regex searches
- Case-sensitive/case-insensitive search toggle
- Pattern highlighting

### Additional Features

- Shell escape with command completion
- Tail -f functionality
- Line numbers
- Session saving
- Help screen

## Usage

### Command Line Options

| <h4>Command Line Option</h4> | <h4>Function</h4>                                |
| :--------------------------- | :----------------------------------------------- |
| -f, --follow                 | follow file changes                              |
| -i, --ignore-case            | search ignores case                              |
| -n, --numbers                | start with line numbers turned on                |
| -p, --pattern                | initial search pattern                           |
| -t, --title                  | page title, default is filename, blank for stdin |
| -v, --version                | print browse version number                      |
| -?, --help                   | print browse command line options                |

### Navigation Commands

| <h4>Pages/Lines</h4>          | <h4>Function</h4>                                  |
| :---------------------------- | :------------------------------------------------- |
| f<br> [PAGE DOWN]<br> [SPACE] | Page down toward EOF                               |
| b<br> [PAGE UP]               | Page up toward SOF                                 |
| ^F<br> ^D<br> z               | Scroll half page down toward EOF                   |
| ^B<br> ^U<br> Z               | Scroll half page up toward SOF                     |
| +<br> [RIGHT]<br> [ENTER]     | Scroll one line toward EOF                         |
| -<br> [LEFT]                  | Scroll one line toward SOF                         |
| d<br> [DOWN]                  | Toggle continuous scroll toward EOF, follow at EOF |
| u<br> [UP]                    | Toggle continuous scroll toward SOF, stop at SOF   |
| ><br> [TAB]                   | Scroll 4 characters right                          |
| <<br> [BACKSPACE]<br> [DEL]   | Scroll 4 characters left                           |
| ^                             | Scroll to column 1                                 |
| $                             | Scroll to EOL                                      |
| 0<br> [HOME]                  | Jump to SOF, column 1                              |
| G                             | Jump to EOF                                        |
| e<br> [END]                   | Jump to EOF, follow at EOF                         |
| t                             | Jump to EOF, tail at EOF                           |

### Search Commands

| <h4>Searches</h4> | <h4>Function</h4>                                                              |
| :---------------- | :----------------------------------------------------------------------------- |
| /                 | Regex search forward, empty pattern repeats search or changes search direction |
| ?                 | Regex search reverse, empty pattern repeats search or changes search direction |
| n                 | Repeat search in the current search direction                                  |
| N                 | Repeat search in the opposite search direction                                 |
| i                 | Toggle between case-sensitive and case-insensitive searches                    |
| p                 | Print the search pattern                                                       |
| P                 | Clear the search pattern                                                       |
| &                 | Run 'grep -nP' on the current file for search pattern                          |

### Miscellaneous Commands

| <h4>Miscellaneous</h4> | <h4>Function</h4>                             |
| :--------------------- | :-------------------------------------------- |
| #                      | Toggle line numbers on and off                |
| %<br> ^G               | Page position                                 |
| !                      | Run a bash command (expands !, %, &)          |
| B                      | Browse another file                           |
| q                      | Quit, save .browserc, next file in list       |
| Q                      | Quit, don't save .browserc, next file in list |
| x                      | Exit list, save .browserc                     |
| X                      | Exit list, don't save .browserc               |

## Configuration

browse saves sessions in `~/.browserc` with the following format:

1. File name
2. First line on page
3. Search pattern
4. Marks
5. Page title

browse saves file history in `~/.browse_files`.

browse saves command history in `~/.browse_shell`.

## Limitations

- Xterm specific
- Logical lines are truncated to screen width
- May be US-centric
- Can be confused by non-printable characters
- Tabs are converted to spaces
- Terminal title changes due to go-prompt dependency

## License

MIT License - see LICENSE file for details
