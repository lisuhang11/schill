# Proposal: Remodel Homepage With LeetCode-Inspired Frontend

## Why

The current homepage already exposes project content, topics, and entry actions, but it still reads like a broad enterprise portal. The requested direction is to reference LeetCode's homepage: a practical first screen that quickly orients users around learning, practice, search, featured content, and progress-oriented discovery.

This project should keep its own visual identity from `web/frontend-style-guide.md`: marine blue-white, light, product-focused, with restrained anime-adjacent warmth. The redesign should borrow LeetCode's information architecture patterns without copying its dark visual theme or proprietary assets.

## What Changes

This change proposes remodeling `web/app/page.tsx` and any narrowly required homepage styling/components so that the homepage:

1. presents a clear top hero with primary actions for starting practice/content discovery,
2. surfaces search as a first-viewport functional entry,
3. adds LeetCode-inspired content modules such as daily challenge, study tracks, topic exploration, featured posts, and progress statistics,
4. keeps existing backend-backed content and topic data where available,
5. uses marine blue-white colors, white surfaces, light blue background, mint/warm accents, and stable 8px-radius cards,
6. preserves responsive behavior across desktop and mobile,
7. avoids introducing unrelated pages, backend endpoints, or speculative features.

## Scope

Included:

- Redesign the home route under `web/app/page.tsx`.
- Add or adjust small route-local components if it reduces homepage complexity.
- Reuse existing API clients such as `getPostList` and `getTopicList`.
- Keep links aligned with existing routes: `/feed`, `/search`, `/topics`, `/posts/new`, `/posts/[postId]`.
- Add LeetCode-inspired modules using project domain language, such as "今日练习", "学习路径", "热门话题", "精选内容", and "社区动态".
- Preserve loading, empty, and error states through existing `StateBlock`, `PostCard`, and `Pagination` where suitable.
- Follow `web/frontend-style-guide.md` visual tokens and layout guidance.

Not included:

- Copying LeetCode source code, CSS, images, text, or brand assets.
- Adding a dark LeetCode clone theme.
- Changing backend service behavior or API contracts.
- Building a full problem bank, online judge, contest system, or study-plan backend.
- Refactoring unrelated frontend pages.
- Replacing the existing app navigation unless homepage integration requires minor text/link alignment.

## Assumptions

- The homepage is the Next.js App Router route at `web/app/page.tsx`.
- Existing route and component conventions should be kept unless the homepage becomes hard to read.
- The style guide is authoritative even though this environment displays parts of it with mojibake; its palette and layout rules are still legible enough to apply.
- LeetCode is used as a product layout reference: fast navigation, practice/discovery modules, progress cues, and content density.
- Any "challenge" or "study path" content can be computed from existing posts/topics for the first version, not backed by new backend endpoints.
- If backend data is unavailable, the homepage should degrade gracefully with clear empty/error states rather than hiding major sections.

## Success Criteria

- The first viewport looks like a usable product homepage, not a marketing landing page.
- Search and primary actions are visible without scrolling on desktop.
- Existing content/topic data still renders when APIs succeed.
- Empty and error states remain visible and readable.
- The page follows the marine blue-white style: light background, white content surfaces, restrained blue, warm/mint accents, and no heavy dark theme.
- Cards, buttons, and panels use stable dimensions and responsive text wrapping.
- The implementation can be verified with lint/typecheck and a browser screenshot after `/opsx:apply`.
