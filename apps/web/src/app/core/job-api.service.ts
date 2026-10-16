import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { Draft, JobPosting, Profile, ResolveJobRequest } from './models/job-posting';

@Injectable({ providedIn: 'root' })
export class JobApiService {
  private readonly http = inject(HttpClient);

  resolveJob(request: ResolveJobRequest): Observable<JobPosting> {
    return this.http.post<JobPosting>('/api/v1/jobs/resolve', request);
  }

  getProfile(): Observable<Profile> {
    return this.http.get<Profile>('/api/v1/profile');
  }

  saveProfile(resumeText: string, background: string): Observable<Profile> {
    return this.http.put<Profile>('/api/v1/profile', { resumeText, background });
  }

  generateDraft(posting: JobPosting): Observable<Draft> {
    return this.http.post<Draft>('/api/v1/drafts', { posting });
  }
}
