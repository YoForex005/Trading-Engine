# Technical Committee - Phase 1: Architect Design
**Topic**: Symbol Search Implementation (Trading Engine)

## Analysis
Currently, the frontend relies on `ws` ticks to populate the "Market Watch" list. This is insufficient for search because:
1.  Symbols with no liquidity/ticks won't appear.
2.  We need a reliable master list to search against.

### Proposed Changes
1.  **Backend (`backend/api/server.go`)**:
    *   Expose `GET /symbols` endpoint.
    *   Source of truth: `s.bbookAPI.GetEngine().GetSymbols()` (via `GetSymbols` method I saw in `engine.go`).
2.  **Frontend (`clients/desktop/src/App.tsx`)**:
    *   Fetch symbol list on mount.
    *   Add a "Search" input field above the symbol list.
    *   Filter the list based on user input.

### Failure Analysis (Mandatory)
1.  **Scenario A: `/symbols` Endpoint Timeout/Fail**
    *   **Behavior**: Fallback to extracting unique symbols from `ticks` (current behavior).
    *   **Log**: `console.warn("Failed to fetch master symbol list")`.
    *   **UX**: User only sees ticking symbols.
2.  **Scenario B: No Matches found**
    *   **Behavior**: Show "No symbols found".
    *   **UX**: Clear feedback.
3.  **Scenario C: Search Input XSS**
    *   **Behavior**: React handles escaping, but avoid `dangerouslySetInnerHTML`.
    *   **Mitigation**: Standard React input binding.

### Functional Implementation
#### Backend
```go
// server.go
func (s *Server) HandleGetSymbols(w http.ResponseWriter, r *http.Request) {
    // Get symbols from engine
    symbols := s.bbookAPI.GetEngine().GetSymbols()
    // Return JSON
}
```

#### Frontend
```tsx
// App.tsx
// Add state
const [searchTerm, setSearchTerm] = useState('');
const [allSymbols, setAllSymbols] = useState<SymbolSpec[]>([]);

// Filter Logic starts with allSymbols, filters by search, updates sortedSymbols
```

### Directives for Coder
-   Implement `HandleGetSymbols` in `backend/api/server.go`.
-   Register the route in `backend/main.go` (Need to check `main.go` first!).
-   Update `App.tsx` with Search Input (Tailwind styled).
-   **Determinism**: Search is case-insensitive.
