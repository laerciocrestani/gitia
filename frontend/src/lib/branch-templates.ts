/** Presets alinhados ao catálogo da TUI (`internal/tui/components/branch_templates.go`). */

export type BranchTemplate = {
  id: string
  prefix: string
  usage: string
  example: string
  other?: boolean
  /** Separador visual após o grupo comum */
  separatorBefore?: boolean
}

const COMMON: BranchTemplate[] = [
  {
    id: "chore",
    prefix: "chore/",
    usage: "Technical tasks without functional impact",
    example: "chore/update-dependencies",
  },
  {
    id: "docs",
    prefix: "docs/",
    usage: "Documentation",
    example: "docs/api-reference",
  },
  {
    id: "feature",
    prefix: "feature/",
    usage: "New feature",
    example: "feature/user-profile",
  },
  {
    id: "fix",
    prefix: "fix/",
    usage: "Bug fix",
    example: "fix/login-error",
  },
  {
    id: "hotfix",
    prefix: "hotfix/",
    usage: "Urgent production fix",
    example: "hotfix/payment-timeout",
  },
  {
    id: "refactor",
    prefix: "refactor/",
    usage: "Refactor without behavior change",
    example: "refactor/auth-service",
  },
  {
    id: "release",
    prefix: "release/",
    usage: "Release preparation",
    example: "release/v2.4.0",
  },
  {
    id: "test",
    prefix: "test/",
    usage: "Tests",
    example: "test/user-controller",
  },
]

const REST: BranchTemplate[] = [
  {
    id: "bugfix",
    prefix: "bugfix/",
    usage: "Bug fix (alternative to fix)",
    example: "bugfix/memory-leak",
    separatorBefore: true,
  },
  {
    id: "build",
    prefix: "build/",
    usage: "Build and tooling",
    example: "build/docker",
  },
  {
    id: "ci",
    prefix: "ci/",
    usage: "CI/CD",
    example: "ci/github-actions",
  },
  {
    id: "develop",
    prefix: "develop",
    usage: "Main development branch (GitFlow)",
    example: "develop",
  },
  {
    id: "experiment",
    prefix: "experiment/",
    usage: "Experiments and POCs",
    example: "experiment/llm-provider",
  },
  {
    id: "main",
    prefix: "main",
    usage: "Production main branch",
    example: "main",
  },
  {
    id: "master",
    prefix: "master",
    usage: "Legacy main branch name",
    example: "master",
  },
  {
    id: "perf",
    prefix: "perf/",
    usage: "Performance improvements",
    example: "perf/query-cache",
  },
  {
    id: "revert",
    prefix: "revert/",
    usage: "Revert changes",
    example: "revert/pr-142",
  },
  {
    id: "spike",
    prefix: "spike/",
    usage: "Technical research",
    example: "spike/openai-responses-api",
  },
  {
    id: "style",
    prefix: "style/",
    usage: "Formatting/style (no logic change)",
    example: "style/php-cs-fixer",
  },
  {
    id: "other",
    prefix: "",
    usage: "Custom branch name",
    example: "my-branch",
    other: true,
  },
]

export const BRANCH_TEMPLATES: BranchTemplate[] = [...COMMON, ...REST]

export function templateNameSeed(t: BranchTemplate): string {
  return t.other ? "" : t.prefix
}

/** Validação básica alinhada à TUI (`validBranchName`). */
export function isValidBranchName(name: string): boolean {
  const n = name.trim()
  if (!n || n.startsWith("-") || n.endsWith("/") || n.includes("..")) return false
  if (n.startsWith(".") || n.endsWith(".")) return false
  for (const ch of n) {
    if (/\s/.test(ch)) return false
    if ("~^:?*[\\@{}".includes(ch)) return false
  }
  return true
}
