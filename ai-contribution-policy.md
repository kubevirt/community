# KubeVirt AI Contribution Policy

## Overview

This policy establishes guidelines for contributions that involve Artificial
Intelligence (AI) tools, including but not limited to Large Language Models
(LLMs), code generation tools, and AI-assisted development environments. This
is a living document that will evolve as AI technology and legal frameworks
mature.

## AI Tool Disclosure Requirements

### Mandatory Disclosure

All contributors **MUST** disclose AI tool use when submitting code,
documentation, or other content to the KubeVirt project.

Disclosure **MUST** take the form of an `Assisted-By` line within the commit
attributing the AI tool used, for example:

```text
Assisted-By: Claude <noreply@anthropic.com>
```

Further details on the type and manner of assistance provided by the tool are
not required but encouraged to give context to reviewers.

Authors should still adhere to KubeVirt's Developer's Certificate of Origin
(DCO) requirements and sign off their commits.

## Acceptable Uses of AI Tools

AI tools are **accepted** as development assistants for:

- **Code scaffolding**: Generating boilerplate code and initial implementations
- **Refactoring**: Suggesting code improvements and modernization
- **Testing**: Creating test cases and test data
- **Documentation**: Drafting technical documentation and code comments
- **Debugging**: Identifying potential issues and suggesting fixes
- **Research**: Exploring architectural approaches and best practices

## Contributor Responsibilities

Contributors using AI tools **MUST**:

1. **Understand completely** all AI-generated code before submission
2. **Conduct thorough human review** of AI-generated content
3. **Ensure compliance** with KubeVirt coding standards and conventions
4. **Verify licensing compatibility** with the Apache 2.0 license
5. **Be capable of debugging** and maintaining the submitted code
6. **Take full responsibility** for the functionality and quality of the
contribution

## Community Perspectives on AI Contributions

### Alternative Approaches

The KubeVirt community recognizes that projects have varying approaches to
AI-generated contributions:

**Restrictive Approach**: Some projects, such as QEMU, have adopted policies to
decline AI-generated contributions entirely. QEMU's position is based on:

- Uncertain copyright and licensing status of AI-generated content
- Potential conflicts with Developer's Certificate of Origin (DCO)
- Legal risks from training materials with restrictive licensing

**Permissive Approach**: Other projects, including those under the Linux
Foundation umbrella, allow AI-generated contributions with proper disclosure
and review.

KubeVirt has chosen a **balanced, disclosure-based approach** that emphasizes
transparency, human oversight, and community review while leveraging AI tools'
productivity benefits.

## Legal and Licensing Considerations

### Copyright Compliance

Contributors must ensure that:

- AI tool terms of service do not conflict with Apache 2.0 licensing
- No copyrighted material is inadvertently included in AI-generated output
- All third-party content is properly attributed and licensed
- The Developer's Certificate of Origin (DCO) can be legitimately signed

### Employer Policies

Contributors should verify that their use of AI tools complies with their
employer's policies regarding AI-generated code in open source contributions.

## Review Process

### Additional Review Requirements

Pull requests with AI-generated content will receive:

- **Standard technical review** following normal KubeVirt processes
- **Additional scrutiny** for code quality, security, and maintainability

### Review Criteria

Reviewers should evaluate:

- Code quality and adherence to project standards
- Appropriate test coverage
- Security implications
- Long-term maintainability

## Policy Violations

Failure to comply with this policy may result in:

- Pull request rejection
- Request for additional disclosure or modification

## Policy Evolution

This policy will be regularly reviewed and updated to reflect:

- Changes in AI technology capabilities
- Legal and regulatory developments
- Community feedback and experience
- Industry best practices

## Questions and Clarifications

For questions about this policy, please:

1. Open an issue in the [community
repository](https://github.com/kubevirt/community)
2. Discuss in the #kubevirt-dev Slack channel or
   <kubevirt-dev@googlegroups.com> mailing list
3. Bring up during community meetings

## References

- [Linux Foundation Generative AI
Guidelines](https://www.linuxfoundation.org/legal/generative-ai)
- [Avocado Framework AI
Policy](https://avocado-framework.readthedocs.io/en/latest/guides/contributor/chapters/ai_policy.html)
- [QEMU Code Provenance
Policy](https://www.qemu.org/docs/master/devel/code-provenance.html#use-of-ai-content-generators)
- [AGENT.md Specification](https://ampcode.com/AGENT.md)
- [KubeVirt Developer's Certificate of
Origin](https://github.com/kubevirt/kubevirt/blob/main/DCO)
