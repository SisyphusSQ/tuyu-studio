import './styles.css'

const capabilities = ['Wails v2 shell', 'Go App facade', 'Frontend asset host']

document.querySelector('#app').innerHTML = `
  <main class="shell">
    <section class="topbar" aria-label="Application status">
      <div>
        <p class="eyebrow">Alpha shell</p>
        <h1>Tuyu Studio</h1>
      </div>
      <span class="status">Desktop shell ready</span>
    </section>

    <section class="workspace" aria-label="Shell preview">
      <aside class="panel">
        <h2>Shell boundary</h2>
        <ul>
          ${capabilities.map((item) => `<li>${item}</li>`).join('')}
        </ul>
      </aside>
      <section class="canvas" aria-label="Workbench mount point">
        <div class="canvas-grid">
          <p class="eyebrow">Workbench mount</p>
          <h2>Canvas-first surface reserved</h2>
          <p>TOO-160 only proves the desktop window and facade boundary. Workbench layout starts in TOO-161.</p>
        </div>
      </section>
      <aside class="panel">
        <h2>Facade</h2>
        <dl>
          <div>
            <dt>Methods</dt>
            <dd>AppInfo, ShellHealth</dd>
          </div>
          <div>
            <dt>Domain access</dt>
            <dd>Not mounted</dd>
          </div>
        </dl>
      </aside>
    </section>
  </main>
`
