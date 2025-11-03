package excel

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/powerlifting-coach-app/program-service/internal/models"
	"github.com/tealeg/xlsx/v3"
)

type ExcelExporter struct{}

func NewExcelExporter() *ExcelExporter {
	return &ExcelExporter{}
}

func (e *ExcelExporter) ExportProgram(program models.Program, sessions []models.TrainingSession, writer io.Writer) error {
	file := xlsx.NewFile()

	// Create program overview sheet
	if err := e.createOverviewSheet(file, program); err != nil {
		return fmt.Errorf("failed to create overview sheet: %w", err)
	}

	// Create weekly breakdown sheets
	if err := e.createWeeklySheets(file, program, sessions); err != nil {
		return fmt.Errorf("failed to create weekly sheets: %w", err)
	}

	// Create exercise database sheet
	if err := e.createExerciseSheet(file, sessions); err != nil {
		return fmt.Errorf("failed to create exercise sheet: %w", err)
	}

	// Write to the provided writer
	return file.Write(writer)
}

func (e *ExcelExporter) createOverviewSheet(file *xlsx.File, program models.Program) error {
	sheet, err := file.AddSheet("Program Overview")
	if err != nil {
		return err
	}

	// Program title
	row := sheet.AddRow()
	cell := row.AddCell()
	cell.Value = program.Name
	cell.GetStyle().Font.Size = 16
	cell.GetStyle().Font.Bold = true

	// Add empty row
	sheet.AddRow()

	// Program details
	details := [][]string{
		{"Program Phase:", string(program.Phase)},
		{"Duration:", fmt.Sprintf("%d weeks", program.WeeksTotal)},
		{"Training Days:", fmt.Sprintf("%d days per week", program.DaysPerWeek)},
		{"Start Date:", program.StartDate.Format("January 2, 2006")},
		{"End Date:", program.EndDate.Format("January 2, 2006")},
		{"AI Generated:", strconv.FormatBool(program.AIGenerated)},
	}

	if program.Description != nil && *program.Description != "" {
		details = append(details, []string{"Description:", *program.Description})
	}

	for _, detail := range details {
		row := sheet.AddRow()
		
		labelCell := row.AddCell()
		labelCell.Value = detail[0]
		labelCell.GetStyle().Font.Bold = true
		
		valueCell := row.AddCell()
		valueCell.Value = detail[1]
	}

	// Add empty row
	sheet.AddRow()

	// Training schedule overview
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.Value = "Training Schedule"
	cell.GetStyle().Font.Size = 14
	cell.GetStyle().Font.Bold = true

	// Weekly overview
	weekMap := make(map[int][]models.TrainingSession)
	for _, session := range e.getSessionsFromProgramData(program) {
		weekMap[session.WeekNumber] = append(weekMap[session.WeekNumber], session)
	}

	for week := 1; week <= program.WeeksTotal; week++ {
		row := sheet.AddRow()
		cell := row.AddCell()
		cell.Value = fmt.Sprintf("Week %d", week)
		cell.GetStyle().Font.Bold = true

		sessions := weekMap[week]
		for _, session := range sessions {
			row := sheet.AddRow()
			row.AddCell().Value = "" // Indent
			
			sessionCell := row.AddCell()
			sessionName := fmt.Sprintf("Day %d", session.DayNumber)
			if session.SessionName != nil {
				sessionName = *session.SessionName
			}
			sessionCell.Value = sessionName
		}
	}

	return nil
}

func (e *ExcelExporter) createWeeklySheets(file *xlsx.File, program models.Program, sessions []models.TrainingSession) error {
	weekMap := make(map[int][]models.TrainingSession)
	for _, session := range sessions {
		weekMap[session.WeekNumber] = append(weekMap[session.WeekNumber], session)
	}

	for week := 1; week <= program.WeeksTotal; week++ {
		sheetName := fmt.Sprintf("Week %d", week)
		sheet, err := file.AddSheet(sheetName)
		if err != nil {
			return err
		}

		// Week header
		row := sheet.AddRow()
		cell := row.AddCell()
		cell.Value = fmt.Sprintf("Week %d Training", week)
		cell.GetStyle().Font.Size = 14
		cell.GetStyle().Font.Bold = true

		sheet.AddRow() // Empty row

		weekSessions := weekMap[week]
		for _, session := range weekSessions {
			if err := e.addSessionToSheet(sheet, session); err != nil {
				return err
			}
			sheet.AddRow() // Empty row between sessions
		}
	}

	return nil
}

func (e *ExcelExporter) addSessionToSheet(sheet *xlsx.Sheet, session models.TrainingSession) error {
	// Session header
	row := sheet.AddRow()
	cell := row.AddCell()
	sessionName := fmt.Sprintf("Day %d", session.DayNumber)
	if session.SessionName != nil {
		sessionName = *session.SessionName
	}
	cell.Value = sessionName
	cell.GetStyle().Font.Bold = true

	// Exercise headers
	row = sheet.AddRow()
	headers := []string{"Exercise", "Sets", "Reps", "Weight (kg)", "RPE", "Rest (sec)", "Notes"}
	for _, header := range headers {
		cell := row.AddCell()
		cell.Value = header
		cell.GetStyle().Font.Bold = true
		cell.GetStyle().Fill.PatternType = "solid"
		cell.GetStyle().Fill.FgColor = "CCCCCC"
	}

	// Add exercises
	for _, exercise := range session.Exercises {
		row := sheet.AddRow()
		
		row.AddCell().Value = exercise.ExerciseName
		row.AddCell().Value = strconv.Itoa(exercise.TargetSets)
		row.AddCell().Value = exercise.TargetReps
		
		weightCell := row.AddCell()
		if exercise.TargetWeightKg != nil {
			weightCell.Value = fmt.Sprintf("%.1f", *exercise.TargetWeightKg)
		} else if exercise.TargetPercentage != nil {
			weightCell.Value = fmt.Sprintf("%.0f%%", *exercise.TargetPercentage)
		}
		
		rpeCell := row.AddCell()
		if exercise.TargetRPE != nil {
			rpeCell.Value = fmt.Sprintf("%.1f", *exercise.TargetRPE)
		}
		
		restCell := row.AddCell()
		if exercise.RestSeconds != nil {
			restCell.Value = fmt.Sprintf("%d", *exercise.RestSeconds)
		}
		
		notesCell := row.AddCell()
		if exercise.Notes != nil {
			notesCell.Value = *exercise.Notes
		}
	}

	return nil
}

func (e *ExcelExporter) createExerciseSheet(file *xlsx.File, sessions []models.TrainingSession) error {
	sheet, err := file.AddSheet("Exercise Database")
	if err != nil {
		return err
	}

	// Sheet header
	row := sheet.AddRow()
	cell := row.AddCell()
	cell.Value = "Exercise Database"
	cell.GetStyle().Font.Size = 14
	cell.GetStyle().Font.Bold = true

	sheet.AddRow() // Empty row

	// Headers
	row = sheet.AddRow()
	headers := []string{"Exercise Name", "Lift Type", "Week", "Day", "Sets", "Reps", "Weight/Intensity", "RPE"}
	for _, header := range headers {
		cell := row.AddCell()
		cell.Value = header
		cell.GetStyle().Font.Bold = true
		cell.GetStyle().Fill.PatternType = "solid"
		cell.GetStyle().Fill.FgColor = "CCCCCC"
	}

	// Add all exercises
	for _, session := range sessions {
		for _, exercise := range session.Exercises {
			row := sheet.AddRow()
			
			row.AddCell().Value = exercise.ExerciseName
			row.AddCell().Value = string(exercise.LiftType)
			row.AddCell().Value = strconv.Itoa(session.WeekNumber)
			row.AddCell().Value = strconv.Itoa(session.DayNumber)
			row.AddCell().Value = strconv.Itoa(exercise.TargetSets)
			row.AddCell().Value = exercise.TargetReps
			
			intensityCell := row.AddCell()
			if exercise.TargetWeightKg != nil {
				intensityCell.Value = fmt.Sprintf("%.1f kg", *exercise.TargetWeightKg)
			} else if exercise.TargetPercentage != nil {
				intensityCell.Value = fmt.Sprintf("%.0f%%", *exercise.TargetPercentage)
			}
			
			rpeCell := row.AddCell()
			if exercise.TargetRPE != nil {
				rpeCell.Value = fmt.Sprintf("%.1f", *exercise.TargetRPE)
			}
		}
	}

	return nil
}

// Helper function to extract sessions from program data
// This would need to be implemented based on your program data structure
func (e *ExcelExporter) getSessionsFromProgramData(program models.Program) []models.TrainingSession {
	// This is a placeholder - you'll need to implement the actual parsing
	// of program data based on how you structure the AI-generated programs
	var sessions []models.TrainingSession
	
	// Example implementation would parse program.ProgramData
	// and create TrainingSession objects
	
	return sessions
}