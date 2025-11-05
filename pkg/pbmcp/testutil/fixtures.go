package testutil

// TestUser represents a test user fixture
type TestUser struct {
	Name  string
	Email string
	Age   int
}

// TestPost represents a test post fixture
type TestPost struct {
	Title     string
	Content   string
	Published bool
}

// GetTestUsers returns test user fixtures
func GetTestUsers() []TestUser {
	return []TestUser{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
		{Name: "Bob", Email: "bob@example.com", Age: 25},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}
}

// GetTestPosts returns test post fixtures
func GetTestPosts() []TestPost {
	return []TestPost{
		{Title: "First Post", Content: "Hello World", Published: true},
		{Title: "Draft Post", Content: "Work in progress", Published: false},
		{Title: "Published Post", Content: "This is live", Published: true},
	}
}

// UserToMap converts TestUser to map for record creation
func UserToMap(user TestUser) map[string]any {
	return map[string]any{
		"name":  user.Name,
		"email": user.Email,
		"age":   user.Age,
	}
}

// PostToMap converts TestPost to map for record creation
func PostToMap(post TestPost) map[string]any {
	return map[string]any{
		"title":     post.Title,
		"content":   post.Content,
		"published": post.Published,
	}
}
