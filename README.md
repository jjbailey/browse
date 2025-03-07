# browse

A simple, unconventional file browser.

## Goals

- Create a file browser with only the most common features
- Keep it simple, keep it friendly

## Features

- Forward and reverse paging
- Forward and reverse scrolling
- Continuous forward and reverse scrolling
- Horizontal scrolling
- Jump to lines
- Mark pages
- Forward and reverse searches by regular expression
- Shell escape
- tail -f
- Line numbers
- Save session
- Help screen

## Scrolling/Following

browse includes several scrolling and following modes:

- Scrolling up and down is a continuous process that provides a seamless browsing experience. Once you start scrolling, browse will continue until you decide to stop it. You can think of the scroll and tail commands as toggle switches.

- When you scroll down and reach the end of the file (EOF), the browse mode switches to "follow" mode. In this mode, it reads and displays two lines of the input file at a time.

- The tail command allows you to jump to and follow EOF, enabling the system to read and display the input file as quickly as possible.

- The cursor position indicates whether the browse mode is following the file. If the cursor is in the lower left-hand corner, browsing is in follow mode. If the cursor is in the upper left-hand corner, browsing is idle.

## Searching

browse utilizes the RE2 regular expression syntax for pattern matching, highlighting all the matches on a line. When browse finds matches not on the visible screen, browse highlights the entire line. Scroll right or left to highlight the match(es).

## Saved Sessions

browse saves sessions in ~/.browserc. The format of the file is plaintext containing the following lines:

1. file name
2. first line on page
3. search pattern
4. marks
5. page title

The session attributes not saved:

- search direction
- numbers
- bash command
- horizontal scroll
- follow/tail mode

When advancing to the next filename in a list of filenames, browse:

- starts at the first page
- resets the horizontal shift to column 1
- turns off follow/tail mode

When browse is called with no filenames, browse attempts to restore the session saved in ~/.browserc.

## Usage

    Usage: browse [-finv] [-p pattern] [-t title] [filename...]
     -f, --follow       follow file
     -i, --ignore-case  search ignores case
     -n, --numbers      line numbers
     -p, --pattern      search pattern
     -t, --title        page title
     -v, --version      print version number
     -?, --help         this message

| <h4>Command Line Option</h4> | <h4>Function</h4>                                |
| :--------------------------- | :----------------------------------------------- |
| -f, --follow                 | follow file changes                              |
| -i, --ignore-case            | search ignores case                              |
| -n, --numbers                | start with line numbers turned on                |
| -p, --pattern                | initial search pattern                           |
| -t, --title                  | page title, default is filename, blank for stdin |
| -v, --version                | print browse version number                      |
| -?, --help                   | print browse command line options                |

<br>

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
| 0<br> [HOME]                  | Jump to line 1, column 1                           |
| G                             | Jump to EOF                                        |
| e<br> [END]                   | Jump to EOF, follow at EOF                         |
| t                             | Jump to EOF, tail at EOF                           |

<br>

| <h4>Jumps/Marks</h4> | <h4>Function</h4>                   |
| :------------------- | :---------------------------------- |
| j                    | Jump to a line                      |
| m                    | Assign top line to mark 1 through 9 |
| 1 - 9                | Jump to marked line, default to SOF |

<br>

| <h4>Searches</h4> | <h4>Function</h4>                                                              |
| :---------------- | :----------------------------------------------------------------------------- |
| /                 | Regex search forward, empty pattern repeats search or changes search direction |
| ?                 | Regex search reverse, empty pattern repeats search or changes search direction |
| n                 | Repeat search in the current search direction                                  |
| N                 | Repeat search in the opposite search direction                                 |
| i                 | Toggle between case-sensitive and case-insensitive searches                    |
| C                 | Clear the search pattern                                                       |
| &                 | Run 'grep -nP' on input file for search pattern                                |

<br>

| <h4>Miscellaneous</h4> | <h4>Function</h4>                     |
| :--------------------- | :------------------------------------ |
| #                      | Toggle line numbers on and off        |
| %<br> ^G               | Page position                         |
| !                      | Run a bash command (expands !, %, &)  |
| B                      | Browse another file                   |
| q                      | Quit, save .browserc, next file       |
| Q                      | Quit, don't save .browserc, next file |
| x                      | Exit, save .browserc                  |
| X                      | Exit, don't save .browserc            |

<br>

## Limitations

- Xterm specific
- Logical lines chopped to the screen width
- Probably US-centric
- Can be confused by lines with non-printable characters
- Tabs mapped to spaces
