# Design: Remodel Homepage With LeetCode-Inspired Frontend

## Overview

The homepage should become a practice-and-discovery dashboard. It should borrow LeetCode's structural strengths:

- clear navigation into problem/content discovery,
- strong search entry,
- visible daily or featured challenge,
- study-plan-like modules,
- content lists with metadata,
- progress and activity cues.

The visual execution must remain project-native: marine blue-white, clean white panels, light blue page background, and small warm or mint accents.

## Current State

`web/app/page.tsx` currently:

- fetches posts and topics on the server,
- renders a large enterprise-style hero,
- shows feature tiles, latest posts, hot topics, and featured posts,
- uses `PostCard`, `Pagination`, `StateBlock`, `getPostList`, `getTopicList`, and `formatCount`.

The page already has useful data wiring. The redesign should preserve that wiring and change the composition.

## Target Information Architecture

### First Viewport

Use a two-column desktop layout and stacked mobile layout:

- Left: greeting/positioning, headline, short supporting copy, primary action buttons, and a search form/link entry.
- Right: "今日练习" or "今日推荐" panel derived from the first available post, with topic badges and metadata.

The first screen should avoid oversized decorative hero content. It should show the actual product actions immediately.

### Practice Hub Row

Add compact modules inspired by LeetCode's homepage areas:

- Daily challenge: links to the top/latest post.
- Study path: links to topics or curated search routes.
- Discussion/community: links to `/feed`.
- Create/publish: links to `/posts/new`.

Each module should be a real navigation entry with icon, label, short status, and stable card size.

### Content Feed

Keep the latest posts list as the main content area. It should:

- retain backend-backed post rendering,
- keep pagination,
- show clear empty/error states,
- use concise section headers and scannable metadata.

### Sidebar

Use a sidebar for:

- popular topics,
- lightweight stats from current page data,
- featured posts.

The sidebar should be useful but secondary. On mobile, it should stack after the main content.

## Visual Direction

Follow `web/frontend-style-guide.md`:

- background: `#EEFBFF` to white,
- primary blue: `#61C9F5`,
- deep blue: `#2F5F9E`,
- panel white: `#FFFFFF`,
- text: `#20303F`,
- muted text: `#687789`,
- warm accent: `#FFD678`,
- mint accent: `#67D8C6`,
- light border: `rgba(77, 100, 124, 0.18)`,
- soft shadow: `0 12px 32px rgba(45, 68, 92, 0.10)`.

Cards should stay around 8px radius. Avoid nested cards unless the inner element is a repeated item such as a post preview or topic pill. Avoid large dark backgrounds, purple-heavy gradients, and decorative-only blobs.

## Data Behavior

Server-side homepage fetch should continue to use:

- `getPostList({ page, pageSize: 10 })`,
- `getTopicList({ page: 1, pageSize: 8, sort: "hot" })`.

Derived homepage data:

- `dailyPost`: first item in the current post list.
- `featuredPosts`: first three posts.
- `topics`: hot topic list from API.
- metrics: total posts, comments on current page, collections on current page, hot topic count.

If posts fail:

- show an error state in the feed area,
- render non-data navigation modules where possible,
- avoid pretending daily challenge data exists.

If posts are empty:

- show an empty state with a publish action,
- allow topic/search/navigation sections to remain visible.

## Component Strategy

Keep the first implementation surgical:

- Prefer route-local helper components inside `web/app/page.tsx` if only used by the homepage.
- Extract a component only when it becomes clearly reusable or the page becomes hard to scan.
- Continue using lucide-react icons already present in the app.
- Preserve existing shared components rather than rewriting `PostCard`, `Pagination`, or `StateBlock`.

Candidate route-local helpers:

- `HeroSearchEntry`
- `DailyChallengePanel`
- `PracticeModule`
- `Metric`
- `TopicPillList`

Only add these if they make the page simpler.

## Accessibility And Responsiveness

- Use semantic headings in order.
- Ensure buttons and links have visible text labels.
- Keep search controls large enough for mobile.
- Make card grids use stable responsive tracks.
- Ensure long Chinese or mixed-language text wraps without overflow.
- Keep first-viewport content usable on mobile without horizontal scrolling.

## Verification Approach

After implementation:

- run `npm run lint` or the configured lint command in `web/`,
- run `npm run typecheck` or `npm run build` if configured and dependencies are available,
- start the local dev server,
- inspect the homepage in the in-app browser at desktop and mobile widths,
- verify screenshot-level requirements: no overlap, no blank primary area, first viewport actions visible, marine blue-white style present.

If a command is blocked by missing dependencies or network restrictions, report the exact command and failure.

## Risks

### Risk 1: LeetCode reference becomes too literal

Mitigation: use LeetCode as an information architecture reference only. Keep project colors, typography, copy, and assets distinct.

### Risk 2: New modules imply unavailable backend capabilities

Mitigation: derive daily/featured/study modules from existing posts and topics, and link to existing routes. Do not add fake problem-solving behavior.

### Risk 3: Homepage grows too dense

Mitigation: prioritize first-viewport search/actions, then content feed. Keep sidebar and module copy short.
