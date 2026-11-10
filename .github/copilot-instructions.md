---
name: synapseos-control
description: "Project skill for SynapseOS Control, a local-first Docker management app (Go backend, React + TypeScript frontend). Covers architecture, Docker security, API, MVP phases, agent working rules, UI/UX (dark minimalist, responsive, Framer Motion), accessibility, copy, and code comments. Load before any change to this repository."
---

# SynapseOS Control: Project Skill

> Primary source of truth for AI agents and developers working on SynapseOS Control. Read it before inspecting, modifying, or implementing any part of the project.

## 0. How to use this file

This file merges four layers. When they conflict, the earlier layer wins:

1. **Project spec** (Part A, sections 1 to 10): what the app is and must never become.
2. **Working rules** (Part B, sections 11 to 14): how to change code safely.
3. **UI, accessibility, copy** (Part C, sections 15 to 17): how the app looks, behaves, and reads.
4. **Code comments** (Part D, section 18): comment hygiene.

If a collision is not resolved here, stop and ask one focused question.

---

# Part A: Project spec

## 1. Project identity

- **Name:** SynapseOS Control
- **Organization:** SynapseOS
- **Type:** Open-source infrastructure management software
- **Category:** Homelab management, Docker management
- **Architecture:** Local-first, remote-ready
- **Backend:** Go
- **Frontend:** React + TypeScript
- **Users:** Homelab enthusiasts, self-hosting users, developers

SynapseOS Control is an independent application in the SynapseOS ecosystem. It gives users one interface to manage Docker containers and monitor local infrastructure, so routine container management does not need the Docker CLI.

It is a standalone product. It must not depend on SynapseOS Agent or any other SynapseOS application.

**Vision:** a simple, reliable, secure interface for managing containerized infrastructure.

**Positioning:** "Your Homelab. Your Infrastructure. Your Control."

## 2. Core development principles

### 2.1 Local-first

- Works locally without internet access after installation.
- The default Docker connection targets the local Docker Engine.
- No mandatory cloud service, external auth provider, or external database service.
- No SynapseOS account required.

### 2.2 Remote-ready

The architecture must allow remote Docker hosts later. Remote management is not part of the MVP.

- Do not build remote networking features during the MVP.
- Do not couple Docker operations to local socket handling. Use a provider interface.
- Future remote connections use secure transports.

### 2.3 Standalone

- Operates independently.
- No mandatory SynapseOS Core, no dependency on SynapseOS Agent, no shared database with other products.
- Optional integrations may come later.

### 2.4 Reliability first

- Use actual Docker data. Never substitute mock data for completed backend functionality.
- Never report a successful operation without checking its result.
- Handle Docker daemon errors gracefully.
- Preserve existing functionality when modifying code.

### 2.5 Incremental development

- One clearly scoped task at a time, each with acceptance criteria and tests.
- Do not start another feature while the current task has unresolved critical issues.

## 3. Technology stack

Use these unless a documented architectural decision changes them.

### Backend

| Technology | Responsibility |
|---|---|
| Go | Primary backend language |
| Docker Engine Go SDK | Docker Engine communication |
| net/http or compatible Go router | HTTP API |
| encoding/json | JSON serialization |
| context | Cancellation and timeouts |
| slog | Structured logging |

Prefer the official Docker Go client, `github.com/moby/moby/client`. Use API version negotiation when the SDK supports it. Do not hand-implement Docker Engine HTTP endpoints the SDK already provides.

### Frontend

| Technology | Responsibility |
|---|---|
| React | UI framework |
| TypeScript | Application language |
| Vite | Build tool |
| Tailwind CSS | Styling |
| shadcn/ui | UI components |
| Framer Motion | UI animations |
| TanStack Query | Server-state management |
| React Router | Client-side routing |
| Recharts | Charts when needed |

Use compatible, stable versions. Add no dependency without a clear technical reason.

### Infrastructure

- Docker Engine
- Docker Compose for application packaging
- SQLite only when persistent configuration or metadata becomes necessary
- Local filesystem for configuration and optional backups

Do not introduce PostgreSQL or Redis into the MVP without a demonstrated requirement.

## 4. System architecture

Three layers: frontend, Go API backend, Docker Engine.

Flow: Browser -> React frontend -> Go API -> Docker provider -> Docker Engine.

The frontend never talks to the Docker socket or Docker Engine API directly. All Docker operations pass through the Go backend.

```text
+--------------------------------+
|          Web Browser           |
|       React + TypeScript       |
+---------------+----------------+
                | HTTP / JSON
                v
+--------------------------------+
|          Go Backend            |
|  HTTP API                      |
|  Application services          |
|  Docker provider               |
|  Local configuration           |
+---------------+----------------+
                | Docker Engine API
                v
+--------------------------------+
|          Docker Engine         |
| Containers / Images / Networks |
+--------------------------------+
```

### Backend responsibilities

Docker connection management, container operations, host information, resource information, API validation, error handling, request cancellation and timeouts, application configuration, security enforcement, structured logging.

Keep Docker operations separate from HTTP handlers. Handlers do not hold large amounts of Docker business logic.

## 5. Docker connection architecture

This is one of the most important decisions in the project. It supports local-first operation and leaves room for remote hosts.

### 5.1 Connection model

- MVP: local Docker host.
- Later: remote Docker host. Do not build it until the local integration is stable.

### 5.2 Docker provider abstraction

Create an interface that separates Docker operations from the connection implementation. Conceptual contract:

```go
type DockerProvider interface {
    Ping(ctx context.Context) error

    Info(ctx context.Context) (HostInfo, error)

    ListContainers(
        ctx context.Context,
        options ContainerListOptions,
    ) ([]Container, error)

    InspectContainer(
        ctx context.Context,
        id string,
    ) (ContainerDetails, error)

    StartContainer(ctx context.Context, id string) error

    StopContainer(ctx context.Context, id string) error

    RestartContainer(ctx context.Context, id string) error

    ContainerLogs(
        ctx context.Context,
        id string,
        options LogOptions,
    ) (io.ReadCloser, error)

    Close() error
}
```

Adapt signatures to the real domain models and SDK. Implement `LocalDockerProvider` for the MVP and `RemoteDockerProvider` later. Do not create a fake `RemoteDockerProvider` that claims to work before remote support exists.

### 5.3 Local connection

Default Linux endpoint: `unix:///var/run/docker.sock`.

The application should detect socket availability, check backend access, ping Docker, retrieve engine information, display connection status, and handle permission errors and an unavailable daemon. Never fall back silently to mock data.

Docker Desktop and non-Linux environments may use other socket paths or configuration. Do not hardcode Linux-only assumptions across the app.

### 5.4 Future remote connection

May use Docker Engine API over TLS, SSH-based connections, or other explicitly supported secure transports.

- Never expose an unauthenticated Docker TCP endpoint.
- Never put Docker credentials in frontend code.
- Never disable TLS certificate verification in production.
- Store credentials securely, use connection timeouts, validate endpoint configuration.
- Make remote connection status visible.

## 6. Docker security requirements

Docker access is highly privileged. Treat it as security-sensitive.

The backend must:

- Never expose the Docker socket to the browser.
- Never accept arbitrary Docker API paths from users.
- Never expose arbitrary command execution through the API.
- Validate container IDs and operation parameters, and restrict operations to supported endpoints.
- Apply request timeouts and handle authorization failures.
- Avoid logging secrets.
- Avoid exposing environment variables or sensitive container configuration by default.

Access to `/var/run/docker.sock` can effectively grant control of the Docker host and may allow host-level privilege escalation. Therefore:

- Document the required permissions clearly.
- Do not recommend world-writable Docker sockets or `chmod 666`.
- Prefer a dedicated deployment configuration with explicitly documented permissions.
- Treat access to the backend as privileged access.

For browser access beyond localhost, add an explicit authentication and network-security design before exposing state-changing operations.

## 7. API design

Versioned API, base path `/api/v1`.

### System

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/health` | Application health |
| GET | `/api/v1/docker/status` | Docker connection status |
| GET | `/api/v1/docker/info` | Docker host information |

### Containers

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/containers` | List containers |
| GET | `/api/v1/containers/:id` | Container details |
| GET | `/api/v1/containers/:id/logs` | Container logs |
| POST | `/api/v1/containers/:id/start` | Start container |
| POST | `/api/v1/containers/:id/stop` | Stop container |
| POST | `/api/v1/containers/:id/restart` | Restart container |

Use one consistent JSON envelope.

```json
{ "data": {}, "error": null }
```

```json
{
  "data": null,
  "error": {
    "code": "DOCKER_CONNECTION_FAILED",
    "message": "Unable to connect to Docker Engine"
  }
}
```

- Use appropriate HTTP status codes.
- Validate path and query parameters, and use request contexts.
- Return useful errors without exposing sensitive internals.
- Document the API.
- Do not implement every endpoint at once.

## 8. MVP scope

The MVP manages one local Docker host.

### Phase 0: Project foundation

Build: repository structure, Go backend init, React frontend init, TypeScript config, shared API types where appropriate, dev scripts, basic logging, environment configuration, README, development docs.

Acceptance: backend starts, frontend starts, frontend reaches the health endpoint, no unnecessary dependencies, structure documented.

### Phase 1: Docker integration

Build: Docker SDK integration, local socket configuration, ping, engine info, connection status, permission error handling.

Acceptance: backend connects to a real local Docker Engine and retrieves its info, correctly reports an unavailable daemon and permission errors, no mock Docker responses.

### Phase 2: Container listing

Build: list containers with names, IDs, images, states, status, and state filtering if practical.

Acceptance: data matches the actual Docker Engine, empty states work, errors display correctly, API responses are validated, the frontend shows real backend data.

### Phase 3: Container details

Build: details, image information, state, creation information, port mappings, network information where appropriate.

Acceptance: details match Docker inspect, missing optional values are handled, sensitive environment data is not exposed by default.

### Phase 4: Container actions

Build: start, stop, restart.

Acceptance: only supported operations are exposed, the backend validates the target container, Docker errors are handled, results are reported honestly, the UI reflects the actual resulting state, destructive actions require appropriate confirmation.

### Phase 5: Container logs

Build: retrieve logs, show stdout and stderr, configurable tail length, handle running and stopped containers, clear loading and error states.

Acceptance: logs come from Docker Engine, large responses are bounded, streaming is handled correctly if implemented, the UI does not freeze on large logs.

### Phase 6: Dashboard

Build: connection status, container count, running and stopped counts, other relevant states, basic resource information where reliably available, recent container information.

Acceptance: every value comes from real backend data, no random or hardcoded production metrics, connection errors are visible, empty states are useful.

### Phase 7: MVP quality assurance

Go tests, Go formatting, static analysis or linting, frontend type checks and linting, API tests, Docker integration tests, UI tests, error-state tests, security review. Do not declare the MVP complete until critical acceptance criteria pass.

## 9. Features outside the initial MVP

Do not implement unless explicitly requested: multi-host management, remote Docker management, Kubernetes, Docker Swarm, container creation and advanced configuration, image building, registry management, Compose editor, browser terminal, Git deployment, automatic deployment, advanced monitoring and historical metrics, alerting, AI assistant, user management, complex RBAC, cloud synchronization.

These may be considered after the local Docker MVP is stable.

## 10. Repository structure

Initial guideline:

```text
synapseos-control/
├── apps/
│   ├── api/
│   │   ├── cmd/
│   │   │   └── server/
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── http/
│   │   │   ├── docker/
│   │   │   ├── services/
│   │   │   └── models/
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── web/
│       ├── src/
│       │   ├── components/
│       │   ├── features/
│       │   ├── hooks/
│       │   ├── lib/
│       │   ├── pages/
│       │   └── types/
│       └── package.json
│
├── docs/
│   ├── architecture.md
│   ├── api.md
│   ├── security.md
│   └── development.md
│
├── deployments/
│   └── docker/
│
├── .gitignore
├── README.md
└── skill.md
```

Adjust only for a concrete technical reason. Do not create empty abstractions or directories that are not needed.

---

# Part B: Working rules

These rules bias toward caution over speed. For trivial tasks, use judgment. AI-generated code can be incorrect, incomplete, or inconsistent, so work is controlled and incremental.

## 11. How to work

### 11.1 Think before coding

Do not assume. Do not hide confusion. Surface tradeoffs.

- State assumptions explicitly. If uncertain, ask.
- If several interpretations exist, present them. Do not pick one silently.
- If a simpler approach exists, say so and push back when warranted.
- If something is unclear, stop, name what is confusing, and ask one focused question.

Before implementing:

1. Read this file and the relevant docs.
2. Inspect the repository and read the relevant source files.
3. Identify the exact task and its acceptance criteria.
4. List the files that may change.
5. Explain the proposed implementation briefly.

Do not modify files before understanding the existing implementation.

### 11.2 Simplicity first

Write the minimum code that solves the problem.

- No features beyond what was asked.
- No abstractions for single-use code. (The `DockerProvider` interface is a documented exception.)
- No configurability that was not requested.
- No error handling for impossible scenarios. Real boundaries (Docker calls, user input, config, socket access) always get error handling.
- If 200 lines could be 50, rewrite it.
- Ask: would a senior engineer call this overcomplicated? If yes, simplify.

### 11.3 Surgical changes

Touch only what you must. Clean up only your own mess.

- Do not "improve" adjacent code, comments, or formatting. Match existing style.
- Do not refactor what is not broken, and do not rewrite unrelated components.
- Mention unrelated dead code instead of deleting it.
- Remove imports, variables, and functions that your own change made unused.
- Do not introduce unapproved architectural changes (see section 13).
- Every changed line traces to the request.

### 11.4 Goal-driven execution

Turn tasks into verifiable goals and loop until verified.

- "Add validation" becomes "write tests for invalid inputs, then make them pass".
- "Fix the bug" becomes "write a test that reproduces it, then make it pass".
- "Refactor X" becomes "tests pass before and after".

For multi-step work, state a short plan:

```text
1. [step] -> verify: [check]
2. [step] -> verify: [check]
3. [step] -> verify: [check]
```

### 11.5 Task boundaries

One task must not implement several unrelated features.

Acceptable: initialize the Go API; add Docker SDK client initialization; implement Docker ping; add the container listing endpoint; connect container listing to the frontend.

Unacceptable: build the entire platform in one operation; implement all container operations, monitoring, and remote management together; rewrite frontend and backend at once without a staged plan.

### 11.6 During implementation

- Follow existing project conventions.
- Do not fabricate API behavior.
- Do not use mock data as production functionality.
- Do not silently ignore errors.
- Do not claim untested functionality works.

### 11.7 After implementation

1. Format Go code (`gofmt`).
2. Run Go tests.
3. Run frontend type checks and linting.
4. Test relevant API endpoints.
5. Verify actual Docker behavior when applicable.
6. Inspect the final diff.
7. Update documentation.
8. Report changed files, actual test results, and remaining limitations.

If a check cannot be run, say why.

These rules are working if diffs hold fewer unnecessary changes, fewer rewrites come from overcomplication, and clarifying questions arrive before implementation instead of after mistakes.

## 12. Testing requirements

Testing is mandatory.

### Backend

- Unit tests for application services and validation.
- HTTP handler tests.
- Docker provider tests.
- Integration tests against a real Docker Engine when available.

Use interfaces and test doubles for isolated unit tests. Doubles are allowed in tests and never replace real Docker integration in production.

### Frontend

- TypeScript checks.
- Component tests where useful, API integration tests where practical.
- Loading, empty, error, and disconnected state tests.
- Responsive layout checks (see 15.3).

### General

Never fabricate test results. If Docker is unavailable in the environment, run the tests that do not need it, name the integration tests that were not executed, and do not claim Docker integration is verified.

## 13. Architecture change policy

The agent must not independently change: backend language, frontend framework, Docker connection architecture, core repository structure, the Docker provider abstraction, the security model, or MVP boundaries.

If a change becomes necessary:

1. Explain the problem.
2. Explain why the current design is insufficient.
3. Present the proposed alternative.
4. Describe the migration impact.
5. Wait for approval before implementing major changes.

Small implementation details may be decided independently when they do not violate this file.

## 14. Definition of done

A feature is complete only when:

- The implementation matches the requested behavior and acceptance criteria.
- Actual data is used where required.
- Error handling exists and security requirements are respected.
- Relevant tests pass and existing functionality remains intact.
- Documentation is updated where necessary.
- Remaining limitations are reported.

A visually complete UI is not proof that the backend works. A successful response from a mock is not proof that Docker integration works. A successful build is not proof that the feature behaves correctly.

---

# Part C: UI, accessibility, copy

## 15. UI and UX

### 15.1 Design direction

A professional developer tool for people who run their own infrastructure.

- **Dark minimalist.** Dark is justified for this audience. Keep one fixed dark theme unless a light theme is requested.
- **Thin glassmorphism,** as a subtle accent only (see 15.2).
- **Cyan and teal accents,** clear typography, consistent spacing.
- **Interactive container cards** (15.5), **tech stack with icons** (15.6), **Framer Motion** (15.4), **responsive mobile-first** (15.3), **smooth scrolling** (15.7).

Avoid: excessive gradients, excessive glow, decoration that reduces usability, fake dashboard metrics, unnecessary animation, overly complex navigation.

### 15.2 Visual rules and dose caps

Defaults are not decisions. Each item below is allowed only at its cap, with a one-line reason.

| Element | Rule |
|---|---|
| Glassmorphism | Thin: low blur, low-opacity fill, 1px border. At most 1 or 2 surfaces at a time (for example the sticky header and an open dialog). Everything else is solid. Text on glass must pass contrast over the worst background behind it. |
| Glow | At most 1 or 2 focus accents. None on every card, icon, or badge. |
| Shadow | Elevation marker only. Most elements sit flat. |
| Border radius | A small fixed set, used deliberately. Not pill-shaped everywhere. |
| Accent color | Cyan is the one accent, used at key moments (primary action, active nav, focus ring). Teal is secondary. Neutrals carry the rest. Cap the palette at 2 or 3 core colors plus 1 accent. |
| Gradients | None as decoration. No blue-to-purple, no rainbow, no blurred hero orbs. Allowed only to separate a hierarchy level, with the reason written down. |
| Background texture | No stock grids, blueprint lines, or dot patterns. |
| Icons | Chosen for meaning. No sparkle, star, or robot icons as feature marks. shadcn/ui ships Lucide by default, which is fine when each icon fits its feature. Pair icons with text labels. |
| Emoji | None in UI text, headings, buttons, or badges. |
| Badges, dots | Only where they mark a real state. No decorative "AI" pills, no eyebrow pill above a headline, no glowing or pulsing dot that marks nothing. |
| Left color stripe | Only where it encodes real state (running, stopped, error). |
| Typography | Choose a typeface for a stated reason. Inter, Geist, and JetBrains Mono are fine with a reason. Monospace is justified for container IDs, image names, ports, and logs. Not for headings as an aesthetic. No wide-tracked all-caps labels as default styling. Keep prose lines under 80 characters. |

#### Starting color tokens

Ratios were computed with the WCAG formula. Re-check any change.

| Token | Hex | Notes |
|---|---|---|
| `--bg` | `#0A0F14` | Page background |
| `--surface` | `#111820` | Cards, panels |
| `--text` | `#E6EDF3` | 16.29:1 on bg, 15.12:1 on surface |
| `--text-muted` | `#8B98A5` | 6.54:1 on bg, 6.07:1 on surface |
| `--accent` | `#22D3EE` | 10.65:1 on bg, 9.89:1 on surface; bg-colored text on accent is also 10.65:1 |
| `--accent-2` | `#14B8A6` | 7.73:1 on bg, 7.18:1 on surface |
| `--border-input` | `#5B6B7C` | 3.27:1 on surface (meets the 3:1 non-text bar) |
| `--danger` | `#F87171` | 6.46:1 on surface |
| `--warning` | `#FBBF24` | 10.70:1 on surface |
| `--success` | `#34D399` | 9.29:1 on surface |

Define them as CSS variables and Tailwind theme tokens. Do not scatter raw hex values through components.

### 15.3 Layout: responsive, mobile-first

Mobile layout is a different layout, not the desktop layout at a smaller size.

- **Mobile-first Tailwind.** Unprefixed classes style the phone. Add `sm:`, `md:`, `lg:` only to enhance.
- **Breakpoints follow content,** not device lists. Narrow the viewport and set the breakpoint where the layout stops working.
- **At least three states:** one column, two columns when a single stack gets too wide, the full grid only when it fits. Check the 600 to 1024 px band explicitly.
- **Mobile scale:** smaller type steps (fluid `clamp()`), tighter section padding and gaps. No desktop padding carried to phones.
- **Grids collapse and stack.** Use `minmax()` or `auto-fit` tracks, not fixed px columns. Use `min-w-0` on grid children.
- **No horizontal scroll.** Container IDs, image names, and port lists wrap or truncate with the full value reachable. Wide tables get an `overflow-x-auto` container, or become stacked cards on phones. Do not hide content with `overflow: hidden`.
- **Heights:** avoid `100vh` sections. Use `auto` or `dvh`.
- **Tap targets:** at least 44 x 44 px, with a clear gap between neighbors. This matters most for start, stop, and restart, which sit side by side.
- **No hover-only interactions.** Every hover reveal has a tap and keyboard equivalent, with visible `:active` feedback.
- **Navigation reflows.** Main navigation is Overview, Containers, Host, Settings. Add an item only when its functionality exists. Desktop: sidebar or top bar. Mobile: a bottom nav (four items fit) or a labeled "Menu" button, never a bare hamburger. Keep fixed bars compact and reserve their height with page padding and `env(safe-area-inset-*)` so the last item stays reachable.
- **Container list:** a table on wide screens, stacked cards on phones. Choose columns from the decision the user makes (is it running, what is it, what can I do), with state and name first.
- **Verify** at a narrow phone width, a tablet width, and desktop, and by dragging the viewport through the range.

### 15.4 Motion with Framer Motion

Motion has a written purpose or it is removed.

- Wrap the app in `<MotionConfig reducedMotion="user">`. Respect `prefers-reduced-motion` in any CSS animation too.
- At most one orchestrated moment per page load (for example a staggered reveal of the Overview summary). Do not fade-and-slide every section.
- Motion answers user action: opening a dialog, expanding a card, switching a view, confirming an action. Use `layout` and `AnimatePresence` for list changes (a container appearing or disappearing from a filtered list).
- Animate to a new container state only after the backend confirms it. Never animate a "running" transition that Docker did not report. While a request runs, show a pending state on the control.
- Durations roughly 150 to 300 ms, eased, no bounce.
- No endless loops: no pulsing, floating, or looping elements without a user trigger. A spinner during a real request is fine.
- Do not stack fade, slide, scale, and float on one element.
- Animate `transform` and `opacity`. Do not animate layout properties on large lists.
- Anything that moves for more than 5 seconds needs a visible pause control.

### 15.5 Interactive container cards

Container and host cards are working controls, not decoration.

- A card has one primary interaction, opening the details route, implemented as a real `<a href>` (React Router `Link`) with a stretched hit area. Quick actions (start, stop, restart) are separate `<button type="button">` elements layered above it. Never a `<div onClick>`, and never a button nested inside a link.
- Content comes from the backend: name, image, state, status, ports. Nothing is invented.
- Show state with text and an icon, never color alone: "Running", "Stopped", "Restarting", "Unhealthy" (use whatever Docker reports). A left edge or dot may add color, but the word carries the meaning.
- Action availability follows real state. Start is disabled for a running container, Stop for a stopped one. A disabled action says why in visible text (for example "Already running"), not only in a hover tooltip.
- Stop and restart ask for confirmation in a dialog (focus trapped, Escape closes, focus returns to the trigger). The dialog states the container name and what will happen.
- During a request: the control shows a pending state and is not clickable twice. On failure: an error message with what Docker said (sanitized) and what to try. On success: the card reflects the backend's refetched state (TanStack Query invalidation).
- States are visible and distinct: rest, hover, `:focus-visible`, active, disabled, pending.
- Hover is subtle (border or surface shift, tiny lift at most). No glow on every card. Cards for different states may differ (an unhealthy container shows its problem). Do not force identical cards where content differs.
- On touch, every hover reveal has a tap equivalent.
- Empty list: say why and what to do ("No containers on this host. Start one with Docker, then refresh").

### 15.6 Tech stack with icons

- Lives in Settings, in an About section. It is one typed constant that mirrors the stack in section 3, not a hand-copied list in JSX.
- Each item: real brand icon, name, one-line role (for example "Go: backend", "Docker Engine SDK: container control"). The name always sits beside the icon.
- Icons are bundled locally (for example SVG components from a package such as `simple-icons` or `react-icons`). No remote icon URLs, so the app still works offline.
- Brand icons are decorative next to a visible label: mark them `aria-hidden="true"`.
- Responsive: a wrapping grid, two columns on phones, more as width allows.
- This describes what the app is built with. It is not a logo bar and implies no endorsement.

### 15.7 Smooth scrolling

- Set `scroll-behavior: smooth` on `html` only inside `@media (prefers-reduced-motion: no-preference)`.
- Set `scroll-padding-top` (and bottom if a bottom nav exists) so anchors and focused elements are not hidden under fixed bars.
- Use native scrolling. No scroll-jacking, momentum libraries, or parallax unless explicitly requested.
- In-page jumps move focus to the target so keyboard and screen-reader users land correctly.
- The log viewer is the exception for scroll behavior: follow-tail scrolls instantly, never smoothly, and can be turned off.

### 15.8 Data honesty and states

Every data-driven page supports **loading, success, empty, error, and disconnected**.

- Loading says what is loading.
- Empty says why it is empty and offers one action.
- Error says what failed and what to do next.
- Disconnected states which host and why (daemon unavailable, permission denied on the socket, timeout), and how to fix it. Permission errors link to the documented permissions, and must never suggest `chmod 666`.
- Never show a success or connected indicator when the Docker daemon is unavailable.
- Every number, count, and row is real. No random or hardcoded production metrics, no invented trend deltas.
- Overview leads with what needs attention (connection problems, unhealthy or stopped containers), then counts. Do not default to the four-stat-cards-plus-chart shell. A chart appears only if it answers a stated question, and that question is its title.
- Every navigation item and control has a real destination or behavior, or a visible "Coming soon" label.
- Do not show sensitive environment variables or secrets in the UI by default.

### 15.9 Logs viewer

- Logs come from Docker. Show stdout and stderr.
- Distinguish stdout and stderr by a text label or icon, not color alone.
- The tail length is configurable and bounded. Cap what the DOM holds (virtualize or trim) so the UI never freezes.
- Monospace is justified here. Long lines wrap or scroll inside their own container, never the page.
- Follow-tail is a visible, keyboard-operable toggle.
- Loading, empty ("No log output yet"), and error states are explicit. Running and stopped containers are both handled.
- The log region is focusable and labeled so keyboard and screen-reader users can scroll and read it.

### 15.10 Accessibility

Accessibility is part of "the UI holds up", not an extra.

- **Contrast:** 4.5:1 for normal text, 3:1 for large text (18px and up) and non-text UI (input borders, focus rings, icons, chart segments). Never claim a pairing passes without computing it. Ratio = (L1 + 0.05) / (L2 + 0.05), with `L = 0.2126 R + 0.7152 G + 0.0722 B` after linearizing each channel (`c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055)^2.4`). Check text over glass at its worst spot.
- **Keyboard:** every control is reachable and operable with Tab, Enter, and Space. Tab order matches visual order. Dialogs trap focus, close on Escape, and return focus. Provide a skip link.
- **Focus:** never `outline: none` without a replacement. Use a visible `:focus-visible` ring (cyan on dark), at least 3:1 against its surroundings.
- **Native elements:** actions are `<button type="button">`, navigation is `<a href>`.
- **Names:** icon-only controls get an `aria-label` for their purpose. Every field has a visible `<label for>`. A placeholder is never the only label.
- **Not color alone:** status, errors, and success carry text or an icon.
- **Live updates:** action results and connection changes are announced (`role="status"` or `aria-live`) without stealing focus.
- **Cursor:** set `cursor: pointer` once in the base layer for enabled interactive elements (Tailwind preflight resets buttons), `cursor: not-allowed` for disabled ones.
- **Zoom:** text resizes to 200% without clipping or horizontal scroll. Use relative units for type.
- **Mobile keyboard:** a focused input scrolls into view and is not covered by the on-screen keyboard.
- **Document semantics:** `<html lang>`, one `<main>`, landmarks, one `<h1>`, no skipped heading levels, a meaningful `<title>`.

### 15.11 Performance

- Give every image, video, and embed intrinsic dimensions or `aspect-ratio`. Reserve space for anything that loads late.
- No render-blocking third-party scripts. The app is local-first, so prefer none.
- Poll or refetch at sensible intervals, and pause polling when the tab is hidden.
- Report only metrics you actually measured.

## 16. Copy and text

UI words exist to help people understand and act.

- **Write from the user's side.** A user manages containers, not "Docker resources". Name things in plain terms.
- **CTAs say what happens:** "Start container", "Restart nginx", "Show logs", "Save settings", not "Submit" or "Click here". An action keeps one name through the flow: "Stop" leads to a toast that says "Stopped".
- **Sentence case** in buttons, headings, labels, and list items. Proper nouns, acronyms, and Docker-reported state names keep their form.
- **Errors do not apologize and are never vague.** Say what happened and how to fix it. Example: "Cannot reach Docker. Check that the Docker daemon is running and that this app can read the socket."
- **Empty screens invite action.** Say why and what fills them.
- **No em dashes or double-hyphen dashes** in text you write. Use a period, comma, colon, or parentheses.
- **No empty AI vocabulary:** unlock, elevate, empower, seamless, robust, cutting-edge, game-changer, revolutionary, journey. Say what the feature does.
- **No fabricated claims,** numbers, names, dates, quotes, or testimonials. Everything is real or a labeled placeholder.
- **No filler patterns:** "In today's fast-paced world", "Let's dive in", "It's not just X, it's Y", forced groups of three, "Let that sink in", "I hope this helps".
- **Name the actor** when known: "You stopped this container", not "The container was stopped".
- **No decorative caps, bold, or quotation marks** for emphasis, and no inline-header bullets ("**Speed:** faster") where a sentence works.
- **Keep the voice plain and direct,** like a colleague writing a note.
- Placeholders that need data say so: `Container name`, `[REAL DATA]`.
- The project tagline in section 1 is the owner's own wording and stays as written.

## 17. Delivery checklist (UI work)

Every line should be true before you report a UI task done.

- [ ] Palette uses the tokens in 15.2, one accent, no default gradients, no decorative glow, grid, or orbs.
- [ ] Glass, glow, shadow, and radius stay at their dose caps.
- [ ] No decorative emoji, generic AI icons, eyebrow pills, or meaningless dots.
- [ ] Layout is mobile-first, has distinct states from phone to desktop, no horizontal scroll, nothing clipped.
- [ ] Tap targets are at least 44 x 44 px with gaps, hover reveals have a tap equivalent, navigation reflows on mobile.
- [ ] Motion has a written purpose, respects reduced motion, has no endless loops, and never shows a state Docker did not report.
- [ ] Container cards use real links and buttons, with visible rest, hover, focus, active, disabled, and pending states.
- [ ] Stop and restart ask for confirmation, and the UI reflects the actual resulting state.
- [ ] Tech stack shows bundled icons with visible names, driven by one typed constant.
- [ ] Smooth scrolling is gated by `prefers-reduced-motion`, anchors are not hidden under fixed bars, log follow-tail is instant.
- [ ] Every text and non-text pairing is verified, focus is visible everywhere, no control relies on color alone.
- [ ] Every data view has loading, empty, error, and disconnected states that name the cause and next action.
- [ ] No connected indicator shows while Docker is unavailable, and no value is mocked or hardcoded.
- [ ] Copy has no em dashes, buzzwords, fabricated claims, or generic CTAs.
- [ ] Go format, Go tests, TypeScript checks, and lint ran, and results are reported honestly.

---

# Part D: Code comments

Comments carry information the code does not already show. Scope guardrail: comment work touches comments only, never executable code, identifiers, imports, formatting, or logic.

## 18. Comment hygiene

Under the surgical-changes rule (11.3), apply this to comments you write or edit as part of your task. Do not mass-strip existing comments across files unless asked.

### Do not write

- Decorative separators and banners (`// ======== DOCKER ========`, ALL CAPS labels).
- Comments that restate the next line or the signature (`// Initialize count` above `count := 0`, `// ListContainers lists containers.`).
- Step narration (`// Step 1: Validate input`, `// First`, `// Next`, `// Finally`).
- Empty labels (`// Main logic`, `// Helper function`, `// Note: this is important`).
- Vague TODOs (`// TODO: improve this`). A TODO names a specific task with enough context to act on.
- Decorative emoji in comments, and end markers (`} // end if`, `// End handleStop`).
- Line-by-line narration. One comment per logical block at most.
- Stiff or loud wording. Use short, sentence-case lines in a natural developer voice.

### Keep and write

Comments that explain business logic and intent, architectural decisions, security considerations, performance trade-offs, concurrency behavior, protocol details, API contracts, workarounds, edge cases and assumptions, licensing and legal notices.

Go note: exported identifiers may need a doc comment that starts with the identifier's name. Make it say something the name does not, such as side effects, errors, or limits. If it can only restate the name, keep it to one short line.

Examples that earn their place here:

```go
// Docker returns 304 when the container is already stopped. Treat it as success.
```

```ts
// Refetch after the action. The backend state is the source of truth, not the click.
```

### Length

A comment is one line, or two when the second carries a new fact. Never three. Keep the constraint (the trap, the silent failure, the rule). Drop the issue number, version history, and reasoning chain. Value does not justify length.

### Comment checklist

- [ ] Every comment adds information the code does not show.
- [ ] No separators, banners, narration, empty labels, vague TODOs, emoji, or end markers.
- [ ] Density is one per logical block, not one per line.
- [ ] Every comment is one line, or two when the second adds a new fact.
- [ ] Only comments changed in comment work, and the code is untouched.

---

# 19. Long-term direction

After the local Docker MVP is stable, SynapseOS Control may expand toward:

1. Remote Docker host management.
2. Multiple host management.
3. Docker Compose management.
4. Image management.
5. Resource monitoring.
6. Host management improvements.
7. Optional integrations with SynapseOS Monitor.
8. Optional integrations with SynapseOS Agent.

These are future directions, not MVP requirements. Every future feature must preserve local-first operation, an independent product architecture, secure Docker communication, maintainability, and evidence-based development.

## Final principle

**Build a reliable Docker management application before building a complete homelab platform.**

Prefer a small, tested, working feature over a large, unverified implementation.

Human-controlled architecture. AI-assisted implementation. Evidence-based verification.

This is the development standard for SynapseOS Control.
