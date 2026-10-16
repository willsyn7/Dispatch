export interface JobPosting {
  source: string;
  companySlug?: string;
  jobId?: string;
  title: string;
  company: string;
  description: string;
  url?: string;
  fetchedAt: string;
}

export interface ResolveJobRequest {
  url?: string;
  title?: string;
  company?: string;
  description?: string;
}

export interface Profile {
  id: string;
  userId: string;
  version: string;
  resumeText: string;
  background?: string;
  updatedAt: string;
}

export interface Draft {
  id: string;
  applicationId: string;
  version: number;
  profileVersion: string;
  coverLetter: string;
  matchSummary: string;
  createdAt: string;
}
