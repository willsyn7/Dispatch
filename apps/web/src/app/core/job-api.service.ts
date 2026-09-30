import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { JobPosting, ResolveJobRequest } from './models/job-posting';

@Injectable({ providedIn: 'root' })
export class JobApiService {
  private readonly http = inject(HttpClient);

  resolveJob(request: ResolveJobRequest): Observable<JobPosting> {
    return this.http.post<JobPosting>('/api/v1/jobs/resolve', request);
  }
}

