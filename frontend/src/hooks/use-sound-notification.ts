// Module-level AudioContext singleton — created during a user gesture (warmUpAudio)
// so it starts in "running" state and can be used from WebSocket event handlers.
let ctx: AudioContext | null = null;
let _soundEnabled = true;

export function setSoundEnabled(v: boolean) { _soundEnabled = v; }
export function isSoundEnabled() { return _soundEnabled; }

function getCtx(): AudioContext {
  if (!ctx) ctx = new AudioContext();
  return ctx;
}

function withCtx(fn: (ctx: AudioContext) => void) {
  if (!_soundEnabled) return;
  try {
    const c = getCtx();
    if (c.state === "suspended") {
      c.resume().then(() => fn(c)).catch(() => {});
    } else {
      fn(c);
    }
  } catch {
    // ignore — AudioContext may be unavailable (e.g. SSR)
  }
}

function playTone(c: AudioContext, freq: number, duration: number, startTime: number) {
  const osc = c.createOscillator();
  const gain = c.createGain();
  osc.connect(gain);
  gain.connect(c.destination);
  osc.frequency.value = freq;
  osc.type = "sine";
  gain.gain.setValueAtTime(0.15, startTime);
  gain.gain.exponentialRampToValueAtTime(0.001, startTime + duration);
  osc.start(startTime);
  osc.stop(startTime + duration);
}

// Call inside a user gesture handler (e.g. join queue) so the AudioContext
// is created and running before WebSocket events arrive.
export function warmUpAudio() {
  try {
    const c = getCtx();
    if (c.state === "suspended") c.resume().catch(() => {});
  } catch {
    // ignore
  }
}

// Soft pop — incoming message
export function playMessageSound() {
  withCtx((c) => {
    const osc = c.createOscillator();
    const gain = c.createGain();
    osc.connect(gain);
    gain.connect(c.destination);
    osc.type = "sine";
    osc.frequency.setValueAtTime(600, c.currentTime);
    osc.frequency.exponentialRampToValueAtTime(400, c.currentTime + 0.08);
    gain.gain.setValueAtTime(0.12, c.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, c.currentTime + 0.08);
    osc.start(c.currentTime);
    osc.stop(c.currentTime + 0.08);
  });
}

// Ascending chime — match found
export function playMatchSound() {
  withCtx((c) => {
    playTone(c, 523, 0.2, c.currentTime);
    playTone(c, 659, 0.25, c.currentTime + 0.2);
  });
}

// Descending tone — partner left
export function playLeaveSound() {
  withCtx((c) => {
    playTone(c, 440, 0.2, c.currentTime);
    playTone(c, 330, 0.25, c.currentTime + 0.2);
  });
}
