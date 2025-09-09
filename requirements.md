Functional Requirements
- Support task creation using a job spec
- Support fetching job execution history
- Support 10k jobs per second
- Support at-most once and at-least once execution
- Support one time and recuring jobs
- Support failures and retry

Optional
- Support task modification and deletion

Non Functional Requirements
- Minimize deviation from the given schedule (Run with 5 second of the scheduled time)
- Highly Available system
- Fault tolerant
- Bounded waiting time



APIs

Job Creation
POST /job
Body
{
    // A modified CRON Spec that includes the Second.
    // The following job is to be executed at
    // Every 31st second of every minute between 01:10 - 01:15 AM every day of the month, every month from Mon-Fri each week
    "schedule" : "31 10-15 1 * * MON-FRI"
    "api": "https://localhost:4444/foo"
    "type": "ATLEAST_ONCE"
}
Response 
{
    jobId
}


Job Deletion
DELETE /job/:jobId => true/false


Job History
GET /job/:jobId/history?limit=10 => JobExecution[]


Entities
Job
JobSchedule
JobExecution
