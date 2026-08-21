import assert from 'node:assert/strict'
import { readdirSync, readFileSync } from 'node:fs'
import { dirname, extname, resolve } from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

import enUS from './en-US.ts'
import koKR from './ko-KR.ts'
import ruRU from './ru-RU.ts'
import viVN from './vi-VN.ts'

type LocaleValue = string | Record<string, unknown> | unknown[]

function collectStrings(value: LocaleValue, path = ''): Array<{ path: string; value: string }> {
  if (typeof value === 'string') return [{ path, value }]
  if (Array.isArray(value)) {
    return value.flatMap((item, index) =>
      collectStrings(item as LocaleValue, `${path}[${index}]`),
    )
  }
  if (value && typeof value === 'object') {
    return Object.entries(value).flatMap(([key, item]) =>
      collectStrings(item as LocaleValue, path ? `${path}.${key}` : key),
    )
  }
  return []
}

function withoutTechnicalTenantTokens(value: string): string {
  return value
    .replace(/\{tenant(?:Id)?\}/g, '')
    .replace(/\btenant_id\b/gi, '')
    .replace(/\btenantless\b/gi, '')
    .replace(/\bX-Tenant-ID\b/g, '')
}

const localeChecks = [
  // Fork Dương Sinh bỏ zh-CN, vi-VN là ngôn ngữ chính (04-nhat-ky-tuy-bien.md C4).
  { name: 'vi-VN', locale: viVN, forbidden: /tenants?/i },
  { name: 'en-US', locale: enUS, forbidden: /\btenants?\b/i },
  { name: 'ko-KR', locale: koKR, forbidden: /테넌트/ },
  { name: 'ru-RU', locale: ruRU, forbidden: /(?:тенант|арендатор)/i },
]

const repositoryRoot = resolve(dirname(fileURLToPath(import.meta.url)), '../../../..')
const publicDocumentationRoots = [
  '.env.example',
  'README_CN.md',
  'client',
  'docker-compose.yml',
  'docs',
  'mcp-server',
]
const documentationExtensions = new Set(['.go', '.json', '.md', '.yaml', '.yml'])

// Thư mục phụ thuộc / sản phẩm build không phải tài liệu công khai: chúng do
// pip/setuptools sinh ra và vẫn mang thuật ngữ cũ của upstream, quét vào chỉ
// tạo báo động giả.
const skippedDirectories = new Set(['.venv', 'node_modules', 'site-packages'])

function collectDocumentationFiles(path: string): string[] {
  const entries = readdirSync(path, { withFileTypes: true })
  return entries.flatMap((entry) => {
    const child = resolve(path, entry.name)
    if (entry.isDirectory()) {
      if (skippedDirectories.has(entry.name) || entry.name.endsWith('.egg-info')) return []
      return collectDocumentationFiles(child)
    }
    return documentationExtensions.has(extname(entry.name)) ? [child] : []
  })
}

test('user-facing locale values use workspace terminology', () => {
  for (const check of localeChecks) {
    const legacyValues = collectStrings(check.locale)
      .filter(({ value }) => check.forbidden.test(withoutTechnicalTenantTokens(value)))
      .map(({ path, value }) => `${check.name}:${path}=${value}`)

    assert.deepEqual(
      legacyValues,
      [],
      `${check.name} still contains user-facing tenant terminology`,
    )
  }
})

test('public documentation uses workspace terminology', () => {
  const legacyLines = publicDocumentationRoots.flatMap((relativePath) => {
    const absolutePath = resolve(repositoryRoot, relativePath)
    const files = extname(absolutePath) ? [absolutePath] : collectDocumentationFiles(absolutePath)
    return files.flatMap((file) =>
      readFileSync(file, 'utf8')
        .split('\n')
        .flatMap((line, index) =>
          line.includes('租户')
            ? [`${file.slice(repositoryRoot.length + 1)}:${index + 1}=${line.trim()}`]
            : [],
        ),
    )
  })

  assert.deepEqual(legacyLines, [], 'public documentation still contains legacy tenant terminology')
})
