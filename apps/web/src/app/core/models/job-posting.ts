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

