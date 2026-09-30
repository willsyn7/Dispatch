import { ChangeDetectionStrategy, Component, OnInit, inject, signal } from '@angular/core';
import { FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { JobApiService } from '../../core/job-api.service';
import { JobPosting } from '../../core/models/job-posting';

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
  readonly mode = signal<'url' | 'manual'>('url');

  readonly form = new FormGroup({
    url: new FormControl('', { nonNullable: true }),
    title: new FormControl('', { nonNullable: true }),
    company: new FormControl('', { nonNullable: true }),
    description: new FormControl('', { nonNullable: true })
  });

  ngOnInit(): void {
    const capturedURL = this.route.snapshot.queryParamMap.get('jobUrl');
    if (capturedURL) {
      this.form.controls.url.setValue(capturedURL);
    }
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
}
