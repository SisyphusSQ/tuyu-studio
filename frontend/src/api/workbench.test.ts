import { describe, expect, it } from 'vitest'

import type { ProjectGraphLayoutSaveCommandDTO } from './dto'
import { DEFAULT_ALPHA_PROJECT_ROOT } from './projectRoot'
import {
  buildProjectGraphViewCommand,
  normalizeProjectGraphLayoutSaveCommand,
  projectOperationRoot,
} from './workbench'

describe('Workbench API root selection', () => {
  it('uses the alpha example project for graph view when no root is selected', () => {
    expect(buildProjectGraphViewCommand(undefined, 'corr-graph')).toMatchObject({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      correlationId: 'corr-graph',
    })
    expect(buildProjectGraphViewCommand(4, 'corr-graph')).toMatchObject({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      expectedGraphVersion: 4,
      correlationId: 'corr-graph',
    })
  })

  it('preserves explicit graph roots for later user-selected projects', () => {
    expect(buildProjectGraphViewCommand({
      root: ' /tmp/tuyu-project ',
      expectedGraphVersion: 7,
    }, 'corr-explicit')).toEqual({
      root: '/tmp/tuyu-project',
      expectedGraphVersion: 7,
      correlationId: 'corr-explicit',
    })
  })

  it('normalizes layout save commands onto the same default graph root', () => {
    const command: ProjectGraphLayoutSaveCommandDTO = {
      root: '',
      expectedGraphVersion: 3,
      viewport: { x: 0, y: 0, zoom: 1 },
      theme: 'dark',
      grid: { visible: true, size: 24, opacity: 0.24 },
      nodes: [],
      correlationId: '',
    }

    expect(normalizeProjectGraphLayoutSaveCommand(command, 'corr-save')).toMatchObject({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      correlationId: 'corr-save',
    })
  })

  it('keeps fixture project actions on the same root while create keeps the cache default', () => {
    expect(projectOperationRoot('create')).toBe('')
    expect(projectOperationRoot('open')).toBe(DEFAULT_ALPHA_PROJECT_ROOT)
    expect(projectOperationRoot('save')).toBe(DEFAULT_ALPHA_PROJECT_ROOT)
    expect(projectOperationRoot('health')).toBe(DEFAULT_ALPHA_PROJECT_ROOT)
  })
})
