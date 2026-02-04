# Commit Message Review Framework

## Overview

This framework establishes a review system for commit message quality
assurance in KubeVirt repositories. The framework improves consistency, reduces
manual review burden, and helps maintain high quality standards as repositories
grow. The framework is implemented in three progressive stages: starting with
commit message linters that provide warnings, progressing to AI-based analysis
that validates commit messages accurately reflect the code changes, and
enhancing the warning system with comprehensive quality checks. All checks
remain as warnings to allow community adaptation.

The framework focuses on ensuring consistent commit message formats, validating
content against template requirements, and providing actionable feedback to
contributors. This is a living document that will evolve as the framework
matures and community feedback is incorporated.

## Motivation

As KubeVirt repositories grow, maintaining consistency across commit messages
becomes increasingly challenging. Manual review processes, while thorough, can
be time-consuming and may miss subtle inconsistencies. Additionally, contributors
may not always be aware of repository-specific conventions and standards.

Current challenges include:

- Inconsistent commit message formats across contributions
- Manual verification of template completeness
- Lack of automated quality checks that can scale with repository growth
- Time spent by reviewers catching common issues that could be automated

An AI-based review framework addresses these challenges by providing automated
checks that can catch issues early, provide consistent guidance to contributors,
and reduce the manual review burden on SIG reviewers and maintainers.

### Framework Goals

The framework aims to:

- Establish consistent commit message format across repositories through
  automated validation
- Provide actionable, clear feedback to contributors on how to fix issues
- Reduce manual review time by catching common issues early
- Complement human reviewers by catching format and template compliance issues
- Allow gradual community adaptation through progressive implementation stages

### Non-Goals

The framework does not:

- Replace human reviewers - automated checks complement human judgment by
  catching format and template compliance issues
- Block PRs - checks remain warnings only to allow
  community adaptation

## Framework Stages

The review framework is implemented in three progressive stages, allowing the
community to adapt gradually and provide feedback at each stage.

### Stage 1: Commit Message Linters

The first stage focuses on establishing consistent commit message formats through
automated linting. Commit message linting tools are integrated via GitHub Actions
to run checks on all commits in pull requests. Initially, these checks display
warnings in PR checklists rather than blocking merges, providing clear error
messages with examples of correct formats.

#### Commit Message Format Exploration

Rather than mandating a specific format immediately, this stage explores options:

- **Conventional Commits**: Widely adopted format with type prefixes (feat, fix,
  docs, etc.)
- **KubeVirt-specific format**: Custom format tailored to KubeVirt project needs
- **Hybrid approach**: Combining elements from multiple formats

The community decides on the preferred format based on:

- Consistency with existing commit history
- Ease of adoption by contributors
- Alignment with KubeVirt project conventions

### Stage 2: AI-Based Analysis

The second stage introduces AI-powered analysis for comprehensive content
validation. A persistent context file (similar to the CLAUDE.md pattern used in
other projects) provides AI systems with repository standards, template
requirements, project-specific instructions, and examples of high-quality commits.
This file serves as a single source of truth for AI analysis, ensuring
consistent evaluation criteria.

#### AI Analysis Capabilities

The AI analysis checks for:

- **Template compliance**: All required sections present and properly formatted
- **Consistency**: Naming conventions, formatting, and style consistency
- **Completeness**: Required links, examples, and documentation present
- **Clarity**: Writing quality and clarity of explanations
- **Accuracy**: Commit message accurately describes the changes made

#### Feedback Mechanism

AI analysis results are displayed in PR comments or checklists, providing
actionable suggestions for each identified issue with links to relevant
documentation and examples.

### Stage 3: Enhanced Warnings

The final stage enhances the warning system with comprehensive quality checks.
All checks remain as warnings displayed in PR checklists and comments, providing
clear feedback without blocking merges. This allows contributors to address
issues while maintaining flexibility for the community. Clear documentation on
standards and how to resolve issues is provided.

## Commit Message Standards

Commit messages should follow the format established by the community during
Stage 1 implementation. While the specific format may vary by repository, all
commit messages should:

- Include a clear, concise subject line (first line)
- Provide a detailed body section when needed
- Reference related issues or PRs when applicable
- Use present tense ("add" not "added")
- Follow repository-specific conventions

### Example Commit Message Formats

#### Conventional Commits Format
```
feat: add AI review framework proposal

Implement automated commit message linting and AI-based content
validation to improve repository quality standards.

Closes #194
```

#### KubeVirt-Specific Format (Example)
```
VEP: Add AI review framework (#194)

Propose integration of AI-based review framework for quality
assurance in enhancements repository.
```

## Implementation Considerations

### Privacy and Security

If using external AI services:

- Data privacy implications are evaluated
- On-premises or self-hosted alternatives are considered when available
- Compliance with project security policies is ensured
- Terms of service for AI providers are reviewed

### Gradual Rollout

The framework is designed for gradual rollout:

- Starting with warnings to allow community adaptation
- Gathering feedback at each stage before progressing
- Adjusting rules based on community input
- Providing grace periods for adoption

### Tool Selection

Specific tools are selected during implementation based on:

- Community preferences and expertise
- Integration capabilities with GitHub
- Maintenance requirements
- Cost considerations
- Open source alternatives availability

## Contributor Responsibilities

Contributors should expect:

- Automated warning checks on their commit messages when opening pull requests
- Clear feedback on any issues found with their commit messages
- Guidance on how to fix issues to meet repository standards

Contributors are encouraged to:

- Review automated feedback and address issues before requesting review
- Provide feedback on the framework to help improve it
- Follow established commit message formats once they are defined
- Ask questions if they need clarification on requirements

## Review Process

### Integration with PR Review

The framework integrates with the standard PR review process:

- Automated checks run on all commits in pull requests
- Results are displayed in PR checklists or comments
- Reviewers can see automated feedback alongside their manual review
- Automated checks complement, but do not replace, human review

### Review Criteria

As with [all contributions to the project](https://github.com/kubevirt/kubevirt/blob/main/docs/reviewer-guide.md),
reviewers should evaluate:

- Commit message quality and adherence to standards
- Content accuracy and completeness
- Overall contribution quality
- Long-term maintainability

Automated checks help reviewers focus on substantive review by catching format
and template compliance issues early.

## Policy Evolution

This framework will be regularly reviewed and updated to reflect:

- Community feedback and experience
- Changes in AI technology capabilities
- Repository growth and evolving needs
- Industry best practices

The framework evolves through community involvement, with feedback collected at
each stage and incorporated into refinements. Rules can be adjusted based on
community input, and the framework can be adapted or rolled back if issues arise.

## Questions and Clarifications

For questions about this framework, please:

1. Open an issue in the [community   repository](https://github.com/kubevirt/community)
2. Discuss in the [#kubevirt-dev](https://kubernetes.slack.com/archives/C0163DT0R8X) Slack channel or
   <kubevirt-dev@googlegroups.com> mailing list
3. Bring up during community meetings

## References

- [Conventional Commits](https://www.conventionalcommits.org/)
- [KubeVirt Developer's Certificate of Origin](https://github.com/kubevirt/kubevirt/blob/main/DCO)
- [KubeVirt Contributor Guide](https://kubevirt.io/user-guide/contributing/)
- [AGENT.md](https://ampcode.com/AGENT.md) standard for AI context files
