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
- tail -f
- Line numbers
- Save session
- Help screen

When filename is absent, browse attmpts to restore the session saved in ~/.browserc.

## Scrolling/Following

browse has several scrolling/following modes.

- Scrolling up and down in browse is a continuous process, providing a seamless browsing experience. Once initiated, scrolling persists until you decide to halt it. Consider the scroll and tail commands as toggle switches.

- When scrolling down hits EOF, browse enters follow mode, reading and displaying two lines per read of the input file.

- The tail command jumps to and follows EOF, reading and displaying the input file as fast as browse can read it.

- The cursor position indicates whether or not browse is following the file. If the cursor is in the lower left-hand corner, browse is following. If the cursor is in the upper left-hand corner, browse is idle.

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
| ^F<br> ^D<br> z | Half page down toward EOF |
| ^B<br> ^U<br> Z | Half page up toward SOF |
| +<br> [RIGHT]<br> [ENTER] | Scroll one line toward EOF |
| -<br> [LEFT] | Scroll one line toward SOF |
| u<br> [UP] | Toggle continuous scroll toward SOF, stop at SOF |
| d<br> [DOWN] | Toggle continuous scroll toward EOF, follow at EOF |
| ><br> | Scroll four characters right |
| <<br> | Scroll four characters left |
| ^<br> [HOME] | Jump to SOF |
| $<br> [END] | Jump to EOF |
| t<br> | Jump to EOF and tail file |
| &nbsp; | &nbsp; |
| <h4>Marks/Jumps</h4> | <h4>Function</h4> |
| j | Jump to a line |
| m | Assign top line to mark 1 through 9 |
| 1 - 9 | Jump to marked line, default to SOF |
| &nbsp; | &nbsp; |
| <h4>Searches</h4> | <h4>Function</h4> |
| / | Regex search forward, empty pattern repeats search or changes search direction |
| ? | Regex search reverse, empty pattern repeats search or changes search direction |
| n | Repeat search in the current search direction |
| N | Repeat search in the opposite search direction |
| i | Toggle case-sensitive and case-insensitive searches |
| C | Clear the search pattern |
| & | Run 'grep -nP' on input file for search pattern |
| &nbsp; | &nbsp; |
| <h4>Miscellaneous</h4> | <h4>Function</h4> |
| # | Toggle line numbers on and off |
| ! | Run a bash command (expands !, %, &) |
| q | Quit browse, save session in ~/.browserc |
| Q | Quit browse, do not save session |
<br>

## Limitations

- Xterm specific
- Logical lines chopped to the screen width
- Searches can find only visible content
- Probably US-centric
- Can be confused by lines with non-printable characters
