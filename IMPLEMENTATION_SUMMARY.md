# AI Coaching System Implementation Summary

## Overview
Complete implementation of an AI-powered powerlifting coaching system with program generation, approval workflows, and workout execution tracking.

## Architecture

### User Flow
```
Registration → Onboarding → AI Chat → Program Review → Approval → Program Overview → Workout Execution
```

### Key Components

#### 1. **Onboarding & Routing** (`OnboardingCheck.tsx`)
- Checks user onboarding status on login
- Routes based on program state:
  - No onboarding → `/onboarding`
  - Onboarded, no program → `/chat`
  - Pending program → `/program` (approval UI)
  - Approved program → `/feed` or `/program` (overview)

#### 2. **AI Chat Interface** (`ChatPage.tsx`, `openwebui.tf`)
- Embedded OpenWebUI iframe for AI coaching conversations
- System prompt includes:
  - Powerlifting federation rules (IPF, USAPL, CPU)
  - Proven programming methodologies (Sheiko, 5/3/1, Calgary Barbell, TSA, Juggernaut)
  - Periodization principles (block, DUP, conjugate)
  - User settings integration (maxes, goals, competition date, training frequency)
- AI returns structured JSON with phases and weeklyWorkouts

#### 3. **Program Approval UI** (`ProgramApprovalView.tsx`)
- Git-style diff interface showing pending program changes
- Displays:
  - Training phases with week ranges and characteristics
  - Weekly workout overview (first 4 weeks preview)
  - Main lifts progression table (squat, bench, deadlift)
- Actions:
  - **Approve**: Moves pending_program_data → program_data, generates workouts
  - **Reject**: Clears pending data, redirects to chat

#### 4. **Program Overview** (`ProgramOverview.tsx`)
- Dashboard showing:
  - Current week and days until competition
  - Phase progress indicators
  - Main lifts progression table with phase annotations
  - Training calendar (card view)
    - Color-coded: Green (completed), Blue (upcoming), Red (missed)
- Week navigation with previous/next buttons

#### 5. **Workout Execution** (`WorkoutDialog.tsx`)
- Interactive workout tracking:
  - Set-by-set logging (weight, reps, RPE)
  - Exercise reordering (drag to reorder)
  - Progress bar showing exercise completion
  - **LocalStorage persistence** - resume workouts after closing app
- Overall workout metrics:
  - Session RPE (1-10)
  - Workout notes
- Visual feedback:
  - Completed sets highlighted in green
  - Progress tracking per exercise

#### 6. **Backend - Program Service**

##### Database Schema
```sql
-- Migration 003: Program approval workflow
ALTER TABLE programs
ADD COLUMN pending_program_data JSONB DEFAULT NULL,
ADD COLUMN program_status VARCHAR(20) NOT NULL DEFAULT 'draft';
-- Status: draft | pending_approval | approved | rejected
```

##### Models (`models.go`)
```go
type ProgramStatus string
const (
    ProgramStatusDraft           ProgramStatus = "draft"
    ProgramStatusPendingApproval ProgramStatus = "pending_approval"
    ProgramStatusApproved        ProgramStatus = "approved"
    ProgramStatusRejected        ProgramStatus = "rejected"
)

type Program struct {
    ProgramData        map[string]interface{}  // Active program
    PendingProgramData *map[string]interface{} // Awaiting approval
    ProgramStatus      ProgramStatus
    // ... other fields
}
```

##### API Endpoints
- `GET /api/v1/programs/active` - Check for active approved program
- `GET /api/v1/programs/pending` - Get pending program for approval
- `POST /api/v1/programs/from-chat` - Create program from AI chat JSON
- `POST /api/v1/programs/:id/approve` - Approve and generate workouts
- `POST /api/v1/programs/:id/reject` - Reject pending changes

##### Workout Generator (`workout_generator.go`)
Parses approved program JSON and creates database entries:

```go
type WorkoutGenerator struct {
    programRepo *repository.ProgramRepository
}

// Parses program_data.weeklyWorkouts and creates:
// - TrainingSession (week, day, scheduled_date)
// - Exercise (name, sets, reps, intensity, RPE)
// - Calculates scheduled dates from program start_date
```

**Input Format (from AI):**
```json
{
  "phases": [
    {
      "name": "Hypertrophy",
      "weeks": [1, 2, 3, 4, 5, 6],
      "focus": "Build work capacity",
      "characteristics": "High volume, moderate intensity (70-80%)"
    }
  ],
  "weeklyWorkouts": [
    {
      "week": 1,
      "workouts": [
        {
          "day": 1,
          "name": "Squat Focus",
          "exercises": [
            {
              "name": "Competition Squat",
              "liftType": "squat",
              "sets": 4,
              "reps": "8",
              "intensity": "70%",
              "rpe": 7,
              "notes": "Focus on depth"
            }
          ]
        }
      ]
    }
  ]
}
```

**Output (Database):**
- `programs` table: Stores program metadata and JSON
- `training_sessions` table: Individual workouts
- `exercises` table: Exercise details for each session
- `completed_sets` table: User logs (weight, reps, RPE, video_id)

#### 7. **Infrastructure** (`openwebui.tf`)
```hcl
# OpenWebUI Deployment
- Kubernetes namespace: openwebui
- Helm chart: open-webui
- ConfigMap: coach-system-prompt (comprehensive coaching instructions)
- Service: ClusterIP on port 80
- Ingress: chat.{domain_name} with TLS
- Integration: LiteLLM endpoint for multi-LLM support

# Variables
- openai_api_key (sensitive)
- litellm_endpoint (default: OpenAI)
- domain_name (for ingress)
```

## Program JSON Schema

### Expected Format from AI
```typescript
interface ProgramData {
  phases: Phase[];
  weeklyWorkouts: WeeklyWorkout[];
  summary: {
    totalWeeks: number;
    trainingDaysPerWeek: number;
    peakWeek: number;
    competitionWeek: number;
  };
}

interface Phase {
  name: string;
  weeks: number[];
  focus: string;
  characteristics: string;
}

interface WeeklyWorkout {
  week: number;
  workouts: Workout[];
}

interface Workout {
  day: number;
  name: string;
  exercises: Exercise[];
}

interface Exercise {
  name: string;
  liftType: 'squat' | 'bench' | 'deadlift' | 'accessory';
  sets: number;
  reps: string; // e.g., "5", "8-10", "AMRAP"
  intensity?: string; // e.g., "70%", "80%"
  rpe?: number; // 1-10
  notes?: string;
  tempo?: string;
  rest?: number; // seconds
}
```

## Features Implemented

### ✅ Core Functionality
1. **User Onboarding**
   - Form completion redirects to chat
   - Settings passed to AI coach

2. **AI Coaching Chat**
   - OpenWebUI integration
   - Custom system prompt with powerlifting knowledge
   - Structured JSON output

3. **Program Approval Workflow**
   - Pending program storage (pending_program_data)
   - Visual diff display
   - Approve/reject actions
   - Automatic workout generation on approval

4. **Program Overview**
   - Current week tracking
   - Phase progress visualization
   - Main lifts progression table
   - Calendar view with workout cards

5. **Workout Execution**
   - Set-by-set tracking
   - Progress persistence (localStorage)
   - Exercise reordering
   - Workout completion logging

6. **Backend Services**
   - Program CRUD operations
   - Approval workflow endpoints
   - Workout generation from JSON
   - Training session management

### 🔄 Partially Implemented
1. **Media Upload for Sets**
   - Database schema ready (video_id in completed_sets)
   - Video upload infrastructure exists
   - UI integration pending

2. **Feed Integration**
   - Video posts show workout context (workout_id field exists)
   - Automatic feed creation on set completion pending

### 📋 Future Enhancements
1. **Workout Analytics**
   - Volume tracking over time
   - RPE trends
   - Progress visualization

2. **Program Adjustments**
   - Mid-program modifications
   - Auto-regulation based on RPE/fatigue

3. **Coach Feedback Integration**
   - Coach reviews AI-generated programs
   - Feedback incorporation into future programs

4. **Advanced Features**
   - Exercise substitutions
   - Deload week auto-insertion
   - Competition day attempt selection

## Database Schema

### Programs Table
```sql
id, athlete_id, coach_id, name, description, phase,
start_date, end_date, weeks_total, days_per_week,
program_data JSONB,           -- Active program
pending_program_data JSONB,   -- Awaiting approval
program_status VARCHAR(20),   -- draft | pending_approval | approved | rejected
ai_generated BOOLEAN,
ai_model, ai_prompt,
is_active, created_at, updated_at
```

### Training Sessions Table
```sql
id, program_id, athlete_id, week_number, day_number,
session_name, scheduled_date, completed_at,
notes, rpe_rating, duration_minutes,
created_at, updated_at
```

### Exercises Table
```sql
id, session_id, exercise_order, lift_type, exercise_name,
target_sets, target_reps, target_weight_kg,
target_rpe, target_percentage, rest_seconds,
notes, tempo, created_at
```

### Completed Sets Table
```sql
id, exercise_id, set_number, reps_completed, weight_kg,
rpe_actual, video_id, notes, completed_at
```

## Testing Checklist

### Manual Testing Flow
1. ✅ Register new user
2. ✅ Complete onboarding form → redirects to /chat
3. ✅ Chat with AI, receive program JSON
4. ✅ Program stored as pending → redirect to /program
5. ✅ Review program in approval UI
6. ✅ Approve program → workouts generated
7. ✅ View program overview with calendar
8. ✅ Start workout → track sets
9. ✅ Close and reopen app → progress persisted
10. ✅ Complete workout → session marked complete

### API Testing
```bash
# Check active program
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8082/api/v1/programs/active

# Create program from chat
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @program.json \
  http://localhost:8082/api/v1/programs/from-chat

# Approve program
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8082/api/v1/programs/{id}/approve
```

## Deployment Instructions

### 1. Apply Database Migrations
```bash
cd services/program-service
# Migrations run automatically on service startup
```

### 2. Deploy OpenWebUI
```bash
cd infrastructure
terraform apply \
  -var="kubernetes_resources_enabled=true" \
  -var="openai_api_key=$OPENAI_API_KEY" \
  -var="domain_name=your-domain.com"
```

### 3. Build and Deploy Services
```bash
# Program service
docker build -t program-service:latest services/program-service
kubectl apply -f k8s/program-service.yaml

# Frontend
cd frontend
npm run build
# Deploy to CDN or static hosting
```

### 4. Configure Environment Variables
```bash
# Frontend (.env)
VITE_OPENWEBUI_URL=https://chat.your-domain.com
REACT_APP_API_URL=https://api.your-domain.com

# Program Service
DATABASE_URL=postgresql://...
RABBITMQ_URL=amqp://...
LITELLM_ENDPOINT=http://litellm:8000/v1
```

## Key Files Modified/Created

### Frontend
- `src/components/Auth/OnboardingCheck.tsx` - Routing logic
- `src/components/Program/ProgramApprovalView.tsx` - Approval UI (new)
- `src/components/Program/ProgramOverview.tsx` - Overview page (new)
- `src/components/Program/WorkoutDialog.tsx` - Workout execution (new)
- `src/pages/ProgramPage.tsx` - Conditional rendering
- `src/pages/ChatPage.tsx` - OpenWebUI integration (new)
- `src/utils/api.ts` - API client methods

### Backend
- `services/program-service/internal/models/models.go` - Program model updates
- `services/program-service/internal/repository/program_repository.go` - New methods
- `services/program-service/internal/handlers/program_handlers.go` - New endpoints
- `services/program-service/internal/services/workout_generator.go` - Generator (new)
- `services/program-service/cmd/main.go` - Wire up generator
- `services/program-service/migrations/003_*.sql` - Schema changes (new)

### Infrastructure
- `infrastructure/openwebui.tf` - OpenWebUI deployment (new)
- `infrastructure/variables.tf` - New variables

## System Prompt Highlights

The AI coach system prompt includes:

1. **Federation Knowledge**: IPF, USAPL, CPU rules and competition procedures
2. **Program Library**: Sheiko, 5/3/1, Calgary Barbell, TSA, Juggernaut, RTS
3. **Periodization**: Block, DUP, Conjugate method principles
4. **Specificity**: Progressive overload, fatigue management, peaking strategies
5. **JSON Format**: Structured output with phases and weeklyWorkouts
6. **Iterative Design**: Users can request modifications before approval

## Success Metrics

✅ **Complete Implementation:**
- User can complete onboarding and be routed to chat
- AI generates program based on user settings
- Program stored as pending for review
- User can approve/reject with visual feedback
- Workouts automatically generated on approval
- Program overview shows comprehensive training plan
- Workout execution with set tracking
- Progress persists across sessions

## Notes

1. **LocalStorage for Progress**: Workouts save progress every update, keyed by `workout-progress-{session.id}`
2. **Phase Detection**: Current phase determined by current week number matching phase.weeks array
3. **Date Calculation**: Scheduled dates computed from program start_date + (week-1)*7 + (day-1)
4. **Main Lifts**: Identified by "competition" keyword in exercise name + lift type match
5. **Media Upload**: Infrastructure ready, UI integration straightforward (add upload button to set rows)

## Conclusion

This implementation provides a complete end-to-end AI coaching system for powerlifting athletes. The system guides users from initial onboarding through AI-assisted program generation, approval workflows, and detailed workout tracking with progress persistence.

The modular architecture allows for easy extension with additional features like analytics, coach feedback integration, and advanced auto-regulation based on athlete performance.
