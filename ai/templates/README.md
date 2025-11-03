# AI Prompt Templates

This directory contains source-of-truth prompt templates for AI coach interactions.

Templates are inserted into the ai_prompt_templates table via database migration at ai-agent-service startup.

## Template Files

### coach_workout_completed.txt
Prompt template for AI coach response after workout completion.
Context includes:
- User settings and goals
- Workout history
- Current program
- Completed workout summary

### coach_dm_response.txt
Prompt template for AI coach direct message responses.
Context includes:
- Conversation history
- User profile
- Recent workouts
- Current program

### coach_program_adjustment.txt
Prompt template for AI-suggested program adjustments.
Context includes:
- Full program
- User settings and goals
- Workout completion history
- Recovery ratings
- Form issues and injuries

## Template Variables

Templates use Go template syntax for variable substitution:
- `{{.UserID}}` - User UUID
- `{{.UserSettings}}` - Full user settings object
- `{{.WorkoutHistory}}` - Array of recent workouts
- `{{.CurrentProgram}}` - Current training program
- `{{.CompDate}}` - Competition date
- `{{.WeeksUntilComp}}` - Calculated weeks until competition
- `{{.ConversationHistory}}` - DM conversation history
- `{{.FormIssues}}` - User-reported form issues
- `{{.Injuries}}` - User-reported injuries
