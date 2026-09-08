# Providing Context to AI Tools

## Documentation-First Approach

KubeVirt adopts an agent-agnostic, documentation-first approach for providing
context to AI tools. Project knowledge - contribution guidelines, coding
patterns, architectural decisions, review policies - lives in standard
documentation files readable by both humans and AI agents.

This approach is preferred over maintaining agent-specific configuration files
(e.g., vendor-specific prompt files or rule sets) for several reasons:

- The community's expertise is in writing good documentation, not in
  maintaining agentic configurations
- Good documentation serves both humans and AI tools simultaneously
- It avoids vendor lock-in to any specific AI tool or platform
- It reduces maintenance burden by having a single source of truth

## Repository Documentation Index

Repositories SHOULD provide a documentation index file that serves as a table
of contents, linking relevant documentation by topic. AI tools can load this
single file to gain awareness of available project documentation and pull in
specific files as needed for the task at hand.

This pattern is adopted by other projects (e.g., Anthropic's
[docs map](https://code.claude.com/docs/en/claude_code_docs_map.md)) and keeps
project knowledge in a form that is useful regardless of which AI tool - or
human - accesses it.

Repository maintainers may choose to additionally maintain an `AGENTS.md`
file, but it is strongly encouraged that it link to a documentation index or
other form of pointers to human-first documentation rather than encoding
project knowledge directly.

## Separation of Documentation and Tool Configuration

When integrating AI-powered tools (e.g., automated code review bots),
repositories SHOULD maintain a clear separation between:

- **Documentation** (agent-agnostic): project knowledge, coding patterns,
  review guidelines, and contribution policies. These live in standard markdown
  files and serve any reader - human or AI.
- **Tool configuration** (vendor-specific): operational settings such as noise
  control, review profiles, and file routing. These live in tool-specific
  configuration files (e.g., `.coderabbit.yaml`) and control only tool
  behavior.

Tool configuration files SHOULD reference documentation files for
project-specific context rather than duplicating knowledge inline.

Repository-level tool configuration applies only to tools that the repository
has officially adopted (e.g., CodeRabbit for automated code review).
Repositories SHOULD NOT include configuration for tools that individual
developers use locally (e.g., Claude Code, Cursor, OpenCode). The choice and
configuration of local development tools is left to each contributor.
