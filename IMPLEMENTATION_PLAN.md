# PowerCoach Feature Implementation Plan

## Overview
This document tracks the systematic implementation of all missing features for the PowerCoach platform.

## Implementation Status

### Epic 1: AI Coach Integration (Priority 1) 🚧 IN PROGRESS
- [x] Infrastructure review
- [ ] Add AI toggle variable to Terraform
- [ ] Create embedded chat interface component
- [ ] Implement chat backend API
- [ ] Add program artifacts system
- [ ] Implement git-like program management
- [ ] Competition date flexibility
- [ ] Post-workout AI analysis

### Epic 2: Enhanced Workout Logging (Priority 2) ⏸️ PENDING
- [ ] Previous set autofill API + UI
- [ ] Warm-up calculator
- [ ] Set tagging system
- [ ] Multi-level notes with media

### Epic 3: Social Feed & Profiles (Priority 3) ⏸️ PENDING
- [ ] Complete Twitter-like feed
- [ ] Athlete profile pages
- [ ] Feed privacy controls

### Epic 4: Coach-Athlete Relationship (Priority 4) ⏸️ PENDING
- [ ] Coach profiles & discovery
- [ ] Coach-athlete connection system
- [ ] Permission revocation system

### Epic 5: Analytics & Visualizations (Priority 5) ⏸️ PENDING
- [ ] Install recharts
- [ ] Muscle heatmap visualization
- [ ] Competition prep dashboard
- [ ] Progress analytics charts

### Epic 6: Workout Management (Priority 6) ⏸️ PENDING
- [ ] Flexible session management
- [ ] Ad-hoc workout support
- [ ] Historical workout management

### Epic 7: Knowledge Base Infrastructure (Priority 7) ⏸️ PENDING
- [ ] Document management system
- [ ] RAG integration with LiteLLM

### Epic 8: UI/UX Polish (Priority 8) ⏸️ PENDING
- [ ] Dark/Light mode
- [ ] Mobile-first responsive design
- [ ] PWA capabilities

## Architecture Decisions

### AI Infrastructure Toggle
- Variable: `ai_features_enabled` in terraform.tfvars
- Controls: LiteLLM deployment, chat API endpoints, AI-related features
- Default: `false` (opt-in for cost control)

### Chat Architecture
- Backend: Go service (program-service) handles all LiteLLM communication
- Frontend: React component using WebSocket for real-time updates
- Library: Custom chat UI (no external dependencies initially)
- Storage: PostgreSQL `ai_conversations` table (already exists)

### Knowledge Base
- Storage: MinIO deployed in Kubernetes
- Vector DB: Qdrant deployed in Kubernetes
- Processing: Background job in program-service

## Database Changes Required

### New Tables
1. `coach_profiles` (coach-service) - Coach information
2. `coach_athlete_relationships` (coach-service) - Relationships with access control
3. `exercise_library` (program-service) - Reusable exercises with metadata
4. `workout_templates` (program-service) - Reusable workout templates
5. `feed_privacy_settings` (video-service) - Privacy controls per post
6. `program_change_log` (program-service) - Git-like change tracking

### Table Modifications
1. `completed_sets` - Add `set_type` enum, `media_urls` JSONB
2. `exercises` - Add `exercise_library_id` FK (nullable)
3. `training_sessions` - Add `is_adhoc` boolean
4. `programs` - Add `competition_date` timestamp

## API Endpoints to Add

### Program Service
- `POST /api/v1/chat/message` - Send chat message to AI
- `GET /api/v1/chat/history` - Get chat history
- `POST /api/v1/programs/changes/propose` - Propose program changes
- `POST /api/v1/programs/changes/apply` - Apply approved changes
- `GET /api/v1/exercises/library` - Get exercise library
- `POST /api/v1/exercises/library` - Add custom exercise
- `GET /api/v1/sessions/history` - Get workout history
- `PUT /api/v1/sessions/:id` - Edit past workout
- `DELETE /api/v1/sessions/:id` - Delete workout (soft delete)
- `POST /api/v1/sessions/:id/clone` - Clone workout as template
- `GET /api/v1/exercises/:name/previous` - Get previous sets for autofill
- `POST /api/v1/exercises/:name/warmup` - Generate warm-up sets
- `GET /api/v1/analytics/volume` - Volume tracking data
- `GET /api/v1/analytics/e1rm` - Estimated 1RM data

### Coach Service
- `POST /api/v1/coaches/profile` - Create/update coach profile
- `GET /api/v1/coaches/search` - Search coach directory
- `POST /api/v1/relationships/request` - Send coaching request
- `POST /api/v1/relationships/:id/accept` - Accept request
- `POST /api/v1/relationships/:id/terminate` - End relationship

### Video Service
- `PUT /api/v1/feed/posts/:id/privacy` - Update post privacy
- `GET /api/v1/feed/athlete/:id` - Get athlete feed (if public)

## Frontend Components to Create

### Chat
- `src/components/Chat/ChatInterface.tsx` - Main chat UI
- `src/components/Chat/MessageList.tsx` - Message display
- `src/components/Chat/MessageInput.tsx` - Input with file upload
- `src/components/Chat/ProgramArtifact.tsx` - Program preview card
- `src/components/Chat/ChangeProposal.tsx` - Git-like diff view

### Analytics
- `src/components/Analytics/AnalyticsDashboard.tsx`
- `src/components/Analytics/VolumeChart.tsx`
- `src/components/Analytics/E1RMChart.tsx`
- `src/components/Analytics/MuscleHeatmap.tsx`
- `src/components/Analytics/CompPrepDashboard.tsx`

### Workout
- `src/components/Workout/WorkoutHistory.tsx`
- `src/components/Workout/WorkoutCalendar.tsx`
- `src/components/Workout/SetLogger.tsx` (enhanced)
- `src/components/Workout/WarmupGenerator.tsx`

### Coach
- `src/components/Coach/CoachProfile.tsx`
- `src/components/Coach/CoachDirectory.tsx`
- `src/components/Coach/RelationshipManager.tsx`

### Profile
- `src/components/Profile/AthleteProfile.tsx`
- `src/components/Profile/CompetitionHistory.tsx`

## Kubernetes Resources to Add

### AI Infrastructure (toggled)
- `k8s/base/qdrant.yaml` - Vector database
- `k8s/base/minio.yaml` - Document storage (or use existing S3)
- Update `k8s/base/litellm.yaml` - Change replicas based on toggle

### Services
- All existing services stay the same
- No new microservices needed (features added to existing services)

## Timeline Estimate

- **Week 1-2**: Epic 1 (AI Chat Integration)
- **Week 3**: Epic 2 (Enhanced Logging)
- **Week 4**: Epic 3 (Social Feed)
- **Week 5**: Epic 4 (Coach-Athlete)
- **Week 6**: Epic 5 (Analytics)
- **Week 7**: Epic 6 (Workout Management)
- **Week 8**: Epic 7 (Knowledge Base)
- **Week 9**: Epic 8 (UI/UX Polish)
- **Week 10**: Testing, bug fixes, optimization

## Next Steps

1. Add `ai_features_enabled` variable to Terraform
2. Create database migrations for all new tables
3. Implement chat interface (frontend + backend)
4. Implement each epic in priority order
5. Test each feature thoroughly before moving to next

---

**Last Updated**: 2025-11-14
**Status**: In Progress - Epic 1
