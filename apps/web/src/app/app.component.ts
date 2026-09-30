import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet],
  template: `
    <header class="site-header">
      <a href="/" class="brand">Job Application Assistant</a>
      <span class="badge">Personal workspace</span>
    </header>
    <main><router-outlet /></main>
  `,
  styles: [`
    .site-header { align-items: center; display: flex; justify-content: space-between; margin: 0 auto; max-width: 72rem; padding: 1.25rem 2rem; }
    .brand { color: var(--ink); font-size: 1.05rem; font-weight: 750; text-decoration: none; }
    .badge { background: var(--surface-raised); border: 1px solid var(--border); border-radius: 999px; color: var(--muted); font-size: .78rem; padding: .35rem .7rem; }
    main { margin: 0 auto; max-width: 72rem; padding: 3rem 2rem 5rem; }
  `],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AppComponent {}

