## Problem Statement

The `c1f` dashboard currently supports basic navigation (`j`/`k`, `Enter`, `Esc`/`b`, `q`), but it doesn't feel like a cohesive vim-style TUI. Terminal power users expect familiar motions like `gg`/`G`, `/` for search, and a `?` help overlay. Without these, navigation feels slower and less natural for users coming from `k9s`, `lazygit`, `vim`, or similar tools.

## Solution

Introduce a focused set of vim-style keybindings across all dashboard views. These bindings will coexist with existing arrow-key and Enter/Esc navigation, making vim the primary scheme while keeping the interface accessible to non-vim users.

## User Stories

1. As a vim user, I want to press `j`/`k` to move up and down lists so I don't have to use arrow keys.
2. As a vim user, I want to press `gg` to jump to the top of a list and `G` to jump to the bottom.
3. As a vim user, I want to press `/` to start filtering/searching the current list.
4. As a vim user, I want to press `?` to see contextual help for the current view.
5. As a non-vim user, I want arrow keys and Enter/Esc to keep working so I'm not forced to learn vim bindings.
6. As a user, I want `Esc` to exit filtering or help mode before navigating back, so keybindings don't conflict.

## Implementation Decisions

- **Default but Coexistent**: Vim bindings are enabled by default. Arrow keys, Enter, Esc, and `q` continue to work as fallbacks.
- **Hierarchical Navigation**: `Enter` (and optionally a future binding) drills down; `Esc`/`b` goes back. `h`/`l` are intentionally reserved for future pane/view switching.
- **List Views**: Workflow List and Instance List use the Bubble Tea `list` component, which already supports `j`/`k`. We will add `gg`/`G` via `list.Top()`/`list.Bottom()` and enable `/` filtering via `l.SetFilteringEnabled(true)` with a custom keybinding trigger.
- **Step Inspector Step List**: The custom step list already supports `j`/`k`. Add `gg`/`G` by resetting the cursor to 0 / len(steps)-1. Add `/` for name-based filtering (using the existing filter infrastructure, but extended to substring search).
- **Help Overlay**: Add a new `help` model/overlay. `?` toggles it per view. It renders global bindings plus view-specific bindings in a scrollable panel.
- **Mode Precedence**: Key handling order is: error overlay → help overlay → filter/search input → normal view bindings. This prevents conflicts.

## Keybinding Map

### Global
| Key | Action |
|-----|--------|
| `q` | Quit |
| `?` | Toggle help overlay |
| `Esc` / `b` | Exit current mode (filter/help/error) or go back |

### Workflow List & Instance List
| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `gg` | Jump to top |
| `G` | Jump to bottom |
| `/` | Start filtering |
| `Enter` | Drill down |

### Step Inspector
| Key | Action |
|-----|--------|
| `j` / `↓` | Move down step list |
| `k` / `↑` | Move up step list |
| `gg` | Jump to first step |
| `G` | Jump to last step |
| `/` | Filter steps by name |
| `f` | Cycle status filter (existing) |
| `Tab` | Switch focus between step list and detail viewport |
| `r` | Refresh instance (existing) |
| `v` | Toggle raw JSON view (existing) |
| `w` | Toggle detail wrapping (existing) |
| `Enter` | (reserved) |

## Out of Scope

- Count prefixes (`5j`, `10G`) — too complex for initial implementation.
- Half-page scrolling (`Ctrl+d`/`Ctrl+u`) — not essential for the core experience.
- `h`/`l` hierarchical navigation — reserved for future pane switching.
- Command-line mode (`:` commands) like `:q` or `:filter`.
- Customizable keybindings or config file support.

## Testing Decisions

- Unit test cursor reset logic for `gg`/`G` in the Step Inspector.
- Unit test that `?` toggles help state and `Esc` dismisses it.
- Verify Bubble Tea `list` filter activation via `/` using model tests.
- Manual QA across all three views to ensure arrow keys still work.

## Further Notes

- Keep the help overlay concise. Group bindings by "Navigation", "Actions", and "View-specific".
- The existing `f` status filter in the Step Inspector should compose cleanly with the new `/` name filter.
- Consider adding a small status hint like "Press ? for help" at the bottom of each view once this feature lands.
