import { ChangeDetectionStrategy, Component, OnInit, inject, signal } from '@angular/core';
import { FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { JobApiService } from '../../core/job-api.service';
import { Draft, JobPosting } from '../../core/models/job-posting';

@Component({
  selector: 'app-job-capture',
  imports: [ReactiveFormsModule],
  templateUrl: './job-capture.component.html',
  styleUrl: './job-capture.component.css',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class JobCaptureComponent implements OnInit {
  private readonly api = inject(JobApiService);
  private readonly route = inject(ActivatedRoute);
  readonly posting = signal<JobPosting | null>(null);
  readonly error = signal('');
  readonly loading = signal(false);
  readonly savingProfile = signal(false);
  readonly generatingDraft = signal(false);
  readonly profileSaved = signal(false);
  readonly draft = signal<Draft | null>(null);
  readonly mode = signal<'url' | 'manual'>('url');

  readonly form = new FormGroup({
    url: new FormControl('', { nonNullable: true }),
    title: new FormControl('', { nonNullable: true }),
    company: new FormControl('', { nonNullable: true }),
    description: new FormControl('', { nonNullable: true })
  });

  readonly profileForm = new FormGroup({
    resumeText: new FormControl('', { nonNullable: true }),
    background: new FormControl('', { nonNullable: true })
  });

  ngOnInit(): void {
    const capturedURL = this.route.snapshot.queryParamMap.get('jobUrl');
    if (capturedURL) {
      this.form.controls.url.setValue(capturedURL);
    }
    this.api.getProfile().subscribe({
      next: profile => {
        this.profileForm.setValue({ resumeText: profile.resumeText, background: profile.background ?? '' });
        this.profileSaved.set(true);
      },
      error: () => undefined
    });
  }

  setMode(mode: 'url' | 'manual'): void {
    this.mode.set(mode);
    this.error.set('');
  }

  submit(): void {
    const value = this.form.getRawValue();
    const request = this.mode() === 'url'
      ? { url: value.url }
      : { title: value.title, company: value.company, description: value.description };

    this.loading.set(true);
    this.error.set('');
    this.api.resolveJob(request).subscribe({
      next: posting => {
        this.posting.set(posting);
        this.loading.set(false);
      },
      error: response => {
        this.error.set(response.error?.error ?? 'The job posting could not be loaded.');
        this.loading.set(false);
      }
    });
  }

  saveProfile(): void {
    const profile = this.profileForm.getRawValue();
    this.savingProfile.set(true);
    this.error.set('');
    this.api.saveProfile(profile.resumeText, profile.background).subscribe({
      next: () => { this.profileSaved.set(true); this.savingProfile.set(false); },
      error: response => { this.error.set(response.error?.error ?? 'The profile could not be saved.'); this.savingProfile.set(false); }
    });
  }

  generateDraft(): void {
    const posting = this.posting();
    if (!posting) return;
    this.generatingDraft.set(true);
    this.error.set('');
    this.api.generateDraft(posting).subscribe({
      next: draft => { this.draft.set(draft); this.generatingDraft.set(false); },
      error: response => { this.error.set(response.error?.error ?? 'The draft could not be generated.'); this.generatingDraft.set(false); }
    });
  }
}
