import { Routes } from '@angular/router';
import { JobCaptureComponent } from './features/job-capture/job-capture.component';

export const routes: Routes = [
  { path: '', component: JobCaptureComponent },
  { path: '**', redirectTo: '' }
];

