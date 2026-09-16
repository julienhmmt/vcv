// Regression test for the shared setup's real-timer cleanup: UI primitives
// (bits-ui Dialog scroll-lock restore) schedule real timers that used to fire
// after a test file's jsdom globals were torn down, failing the run with an
// unhandled "document is not defined".
import { describe, expect, it } from 'vitest'

let leakedTimerFired = false

describe('test timer cleanup', () => {
  it('schedules a real timer that must never fire', () => {
    setTimeout(() => {
      leakedTimerFired = true
    }, 20)
  })

  it('drops timers left pending by earlier tests', async () => {
    await new Promise((resolve) => setTimeout(resolve, 100))
    expect(leakedTimerFired).toBe(false)
  })
})
