# DESIGN.md — Product design system (entry point)

> Entry point: principles, tokens, navigation across the library.
> Components live in design-guide/, page templates in template/.
> The source of truth (UI framework/template) and brand values — fill in once chosen.
> For AI: load at the start of a design session; do not reconstruct the system from scratch.

## How to use
- Before laying out a page → the component catalog and the list of available libraries.
- Before changing styles → the tokens and "Principles and rules". Color/spacing via tokens, not hardcoded.

## Principles and rules
Do: reuse components; color/spacing/radii via tokens; custom work as a separate layer.
Don't: don't edit compiled static assets; don't spawn colors outside the palette.

## Tokens
- Color: palette and roles (background, text, accent, states).
- Typography: fonts, scale, weights.
- Spacing and grid.

## Library navigation
Descriptive layer — tokens, themes, layouts, icons.
Component layer — catalog, component cards, ready-made snippets.

## Accessibility
Contrast, focus, keyboard, alt text — required.
