# Tasks: Remodel Homepage With LeetCode-Inspired Frontend

## 1. Confirm Homepage Inputs

- [x] Re-read `web/app/page.tsx` and identify current data dependencies.
- [x] Re-read `web/frontend-style-guide.md` for applicable palette, spacing, radius, and homepage guidance.
- [x] Review the provided LeetCode homepage reference and extract only reusable layout patterns.
- [x] Confirm existing routes used by homepage links: `/feed`, `/search`, `/topics`, `/posts/new`, and `/posts/[postId]`.

## 2. Redesign Homepage Structure

- [x] Replace the enterprise portal hero with a practice/discovery first viewport.
- [x] Add a visible search entry in the first viewport.
- [x] Add a daily challenge or daily recommendation panel derived from existing post data.
- [x] Add compact practice/discovery modules for daily content, study paths, community feed, and publishing.
- [x] Preserve latest posts as the main backend-backed content list.
- [x] Preserve hot topics, featured posts, and metrics in a secondary area.

## 3. Preserve Data And States

- [x] Keep `getPostList` and `getTopicList` server-side data fetching.
- [x] Keep pagination behavior for the latest posts list.
- [x] Keep error state rendering when posts fail to load.
- [x] Keep empty state rendering when no posts exist.
- [x] Add graceful fallbacks when daily/featured content is missing.

## 4. Apply Marine Blue-White Style

- [x] Use the style guide palette: white panels, light blue background, deep blue text/actions, warm and mint accents.
- [x] Keep cards around 8px radius and avoid decorative-heavy nested panels.
- [x] Use lucide-react icons for module buttons and navigation cues.
- [x] Ensure text, buttons, and cards do not overflow on mobile.
- [x] Avoid dark LeetCode clone styling, large purple gradients, or decorative blobs.

## 5. Verify

- [x] Run the configured frontend lint command from `web/`.
- [x] Run the configured typecheck or build command from `web/`.
- [x] Start the local frontend dev server if needed.
- [ ] Inspect the homepage in the in-app browser at desktop width.
- [ ] Inspect the homepage in the in-app browser at mobile width.
- [ ] Confirm first-viewport search and primary actions are visible.
- [x] Record any blocked verification commands in the implementation summary.
