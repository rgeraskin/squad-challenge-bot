package domain

// Business logic limits - adjust these values to configure system constraints
const (
	// MaxChallengesPerUser is the maximum number of challenges a user can join
	MaxChallengesPerUser = 10

	// MaxParticipants is the maximum number of participants allowed per challenge
	MaxParticipants = 50

	// MaxTasksPerChallenge is the maximum number of tasks allowed per challenge
	MaxTasksPerChallenge = 50

	// MaxTaskDescriptionLength is the maximum character length for task descriptions
	MaxTaskDescriptionLength = 1200
)
