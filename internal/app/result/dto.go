package result

// This domain only exposes read endpoints. The response shapes
// (result.EventResult, result.FinalizeReadiness, result.LeaderboardSchoolScore)
// already carry json tags and are returned directly — the same convention
// score.EventScoresDetailed uses — so no separate response DTOs are needed
// here.
