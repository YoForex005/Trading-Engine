---
description: Run a rigorous 'Technical Committee' coding process (Architect -> Coder -> Reviewer).
---

# Technical Committee Protocol

Follow these steps for every coding task when this workflow is active.

---

## ⚙️ Configuration

**Model Assignment:**
- **Gemini 3 Pro**: Architect & Presenter (Phases 1, 4)
- **Claude Sonnet 4.5 (Thinking)**: Implementation (Phase 2)
- **Claude Opus 4.5 (Thinking)**: Security Review & QA (Phase 3)

**Iteration Limits:**
- Maximum review-fix cycles: **2**
- If no critical/major issues remain after 2 cycles, proceed to Phase 4.

**Security Scope (Assume by default):**
- High concurrency environments
- Malicious users attempting exploitation
- Financial/data risk if vulnerabilities exist

---

## Phase 1: The Architect (Gemini 3 Pro)
**Model**: Gemini 3 Pro
**Goal**: Design a robust, secure, and scalable solution.

1.  Analyze the User Request thoroughly.
2.  Identify potential edge cases, security risks, and architectural trade-offs.
3.  **Failure Analysis (Mandatory):**
    - List at least **3 distinct failure scenarios**.
    - Define system behavior under partial failure.
    - Explicitly state what is **logged** (internal) vs what is **returned** to the user (safe error messages).
4.  Outline the file structure and key function signatures *before* writing any code.
5.  **Output**: A mini-spec or implementation plan.

---

## Phase 2: The Coder (Claude Sonnet 4.5 – Thinking)
**Model**: Claude Sonnet 4.5 (Thinking)
**Goal**: Implement the Architect's spec with precision and efficiency.

1.  Write the actual code based on the Phase 1 spec.
2.  **Determinism Rule**: Avoid non-deterministic behavior unless explicitly required. No hidden randomness, time-based logic, or implicit globals.
3.  Focus on clean, readable, and idiomatic code.
4.  Do not cut corners on error handling.
5.  Include strict type hints (TypeScript/Python).

> [!IMPORTANT]
> Claude Sonnet writes ALL code. Gemini/Opus do NOT implement.

---

## Phase 3: The Reviewer (Claude Opus 4.5 – Thinking)
**Model**: Claude Opus 4.5 (Thinking)
**Goal**: Ruthlessly critique and attack the code.

1.  **Stop**: Do not show the code to the user yet.
2.  **Design Freeze**: You may **not** introduce new features or redesign the architecture. Suggestions must be limited to correctness, security, and robustness fixes.
3.  **Review with Paranoid Assumptions**:
    - Vulnerabilities (Injection, XSS, CSRF)
    - Race conditions / Concurrency
    - Logic bugs
4.  **Issue Classification (Required)**:
    Each issue must be labeled:
    - **[CRITICAL]** (Must fix: Security holes, data loss)
    - **[MAJOR]** (Should fix: Logic errors, scale issues)
    - **[MINOR]** (Optional: Style, comments)
    *Only CRITICAL and MAJOR issues trigger a rewrite/fix cycle.*
5.  **Evidence-Based Feedback**:
    > [!CAUTION]
    > All criticisms **MUST**:
    > - Reference **specific lines or functions**
    > - Explain the **real-world impact**
    > - Be **actionable**

---

## Phase 4: Final Presentation (Gemini 3 Pro High)
**Model**: Gemini 3 Pro High
**Goal**: Deliver the polished artifact.

1.  Present the final, polished code to the user.
2.  **Diff Summary**:
    - **Changes**: What changed from the initial implementation?
    - **Bugs Caught**: What specific bugs did the Reviewer find?
    - **Safety**: Why is this final version safer/better?

---

## 📌 Summary Table

| Phase | Role       | Model                        | Responsibility                     |
|-------|------------|------------------------------|------------------------------------|
| 1     | Architect  | Gemini 3 Pro                 | Design, failure analysis, API spec |
| 2     | Coder      | Claude Sonnet 4.5 (Thinking) | Deterministic implementation       |
| 3     | Reviewer   | Claude Opus 4.5 (Thinking)   | Classified critique (Crit/Maj/Min) |
| 4     | Presenter  | Gemini 3 Pro High            | Final code & Diff Summary          |
