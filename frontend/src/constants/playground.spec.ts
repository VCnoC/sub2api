import { describe, expect, it } from 'vitest'
import {
  getGptImage2AspectRatio,
  getGptImage2Resolution,
  getGptImage2ResolutionForSize,
  getGptImage2Size,
  isGptImageModel,
  withGptImage2Resolution,
} from './playground'

describe('playground image controls', () => {
  it('recognizes image model suffixes and keeps resolution in the model name', () => {
    expect(isGptImageModel('gpt-image-2-4k')).toBe(true)
    expect(isGptImageModel('image-2-4k')).toBe(true)
    expect(getGptImage2Resolution('gpt-image-2-4k')).toBe('4K')
    expect(getGptImage2Resolution('image-2-4k')).toBe('4K')
    expect(getGptImage2Resolution('nano-banana2-4k')).toBeNull()
    expect(withGptImage2Resolution('gpt-image-2-1k', '2K')).toBe('gpt-image-2-2k')
    expect(withGptImage2Resolution('image-2-1k', '2K')).toBe('image-2-2k')
    expect(withGptImage2Resolution('gpt-image-2-vip', '2K')).toBe('gpt-image-2-vip')
  })

  it('maps aspect ratio and resolution to an exact upstream size', () => {
    expect(getGptImage2Size('2K', '16:9')).toBe('2560x1440')
    expect(getGptImage2ResolutionForSize('2560x1440')).toBe('2K')
    expect(getGptImage2AspectRatio('2560x1440')).toBe('16:9')
  })
})
