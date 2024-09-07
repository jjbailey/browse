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

browse has several scrolling/following modes:

- Scrolling up and down in browse is a continuous process, providing a seamless browsing experience.  Once initiated, scrolling persists until you decide to halt it.  Consider the scroll and tail commands as toggle switches.

- When scrolling down hits EOF, browse enters follow mode, reading and displaying two lines per read of the input file.

- The tail command jumps to and follows EOF, reading and displaying the input file as fast as browse can read it.

- The cursor position indicates whether or not browse is following the file.  If the cursor is in the lower left-hand corner, browse follows.  If the cursor is in the upper left-hand corner, browse is idle.

## Searching

browse utilizes the RE2 regular expression syntax for pattern matching, highlighting all the matches on a line.  When browse finds matches not on the visible screen, browse highlights the entire line.  Scroll right or left to highlight the match(es).

## Saved Sessions

browse saves sessions in ~/.browserc.  The format of the file is plaintext containing the following lines:

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

When a filename is absent, browse attempts to restore the session saved in ~/.browserc.

## Usage

    Usage: browse [-finv] [-p pattern] [-t title] [filename]
     -f, --follow       follow file
     -i, --ignore-case  search ignores case
     -n, --numbers      line numbers
     -p, --pattern      search pattern
     -t, --title        page title
     -v, --version      print version number
     -?, --help         this message

| <h4>Command Line Option</h4> | <h4>Function</h4> |
| :-- | :-- |
| -f, --follow | follow file changes |
| -i, --ignore-case | search ignores case |
| -n, --numbers | start with line numbers turned on |
| -p, --pattern | initial search pattern |
| -t, --title | page title, default is filename, blank for stdin |
| -v, --version | print browse version number |
| -?, --help | print browse command line options |
<br>

| <h4>Pages/Lines</h4> | <h4>Function</h4> |
| :-- | :-- |
| f<br> [PAGE DOWN]<br> [SPACE] | Page down toward EOF |
| b<br> [PAGE UP] | Page up toward SOF |
| ^F<br> ^D<br> z | Scroll half page down toward EOF |
| ^B<br> ^U<br> Z | Scroll half page up toward SOF |
| +<br> [RIGHT]<br> [ENTER] | Scroll one line toward EOF |
| -<br> [LEFT] | Scroll one line toward SOF |
| d<br> [DOWN] | Toggle continuous scroll toward EOF, follow at EOF |
| u<br> [UP] | Toggle continuous scroll toward SOF, stop at SOF |
| ><br> [TAB] | Shift four characters right |
| <<br> [BACKSPACE] | Shift four characters left |
| ^<br> [HOME] | Jump to SOF |
| G<br> $ | Jump to EOF |
| e<br> [END] | Jump to EOF, follow at EOF |
| t | Jump to EOF, tail at EOF |
| &nbsp; | &nbsp; |
| <h4>Jumps/Marks</h4> | <h4>Function</h4> |
| j | Jump to a line |
| m | Assign top line to mark 1 through 9 |
| 0 (zero) | Jump to line 1, shift to column 1 |
| 1 - 9 | Jump to marked line, default to SOF |
| &nbsp; | &nbsp; |
| <h4>Searches</h4> | <h4>Function</h4> |
| / | Regex search forward, empty pattern repeats search or changes search direction |
| ? | Regex search reverse, empty pattern repeats search or changes search direction |
| n | Repeat search in the current search direction |
| N | Repeat search in the opposite search direction |
| i | Toggle between case-sensitive and case-insensitive searches |
| C | Clear the search pattern |
| & | Run 'grep -nP' on input file for search pattern |
| &nbsp; | &nbsp; |
| <h4>Miscellaneous</h4> | <h4>Function</h4> |
| # | Toggle line numbers on and off |
| %<br> ^G | Page position |
| ! | Run a bash command (expands !, %, &) |
| q | Quit browse, save session in ~/.browserc |
| Q | Quit browse, do not save session |
<br>

## Limitations

- Xterm specific
- Logical lines chopped to the screen width
- Probably US-centric
- Can be confused by lines with non-printable characters
