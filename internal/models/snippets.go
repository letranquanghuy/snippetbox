package models

import (
	"database/sql"
	"errors"
	"time"
)

// Define a Snippet type to hold the data for an individual snippet. Notice how
// the fields of the struct correspond to the fields in our MySQL snippets
// table?
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// Define a SnippetModel type which wraps a sql.DB connection pool.
type SnippetModel struct {
	DB *sql.DB
}

// This will insert a new snippet into the database
func (m *SnippetModel) Insert(title, content string, expires int) (int, error) {
	// Write the SQL statement we want to execute. I've split it over two lines
	// for readability (which is why it's surrounded with backquotes instead
	// of normal double quotes).
	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// Use the Exec() method on the embedded connection pool to execute the
	// statement. The first parameter is the SQL statement, followed by the
	// values for the placeholder parameters: title, content and expiry in
	// that order. This method returns a sql.Result type, which contains some
	// basic information about what happened when the statement was executed
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	// Use the LastInsertId() method on the sql.Result to get the ID of the
	// newly inserted snippet. This is important because we'll need this ID
	// later when we want to display the snippet.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// The ID returned from LastInsertId() is an int64, but our Snippet type
	// uses an int for the ID field. So we convert it to an int before
	// returning.
	return int(id), nil
}

// // This will return a specific snippet based on its ID.
// func (m *SnippetModel) Get(id int) (Snippet, error) {
// 	// Write the SQL statement we want to execute. Again, I've split it over two
// 	// lines for readability.
// 	stmt := `SELECT id, title, content, created, expires FROM snippets
// 	WHERE expires > UTC_TIMESTAMP() AND id = ?`

// 	// Use the QueryRow() method on the connection pool to execute our
// 	// SQL statement, passing in the untrusted id variable as the value for the
// 	// placeholder parameter. This returns a pointer to a sql.Row object which
// 	// holds the result from the database.
// 	row := m.DB.QueryRow(stmt, id)

// 	// Initialize a new zeroed Snippet struct.
// 	var s Snippet

// 	// Use row.Scan() to copy the values from each field in sql.Row to the
// 	// corresponding field in the Snippet struct. Notice that the arguments
// 	// to row.Scan are *pointers* to the place you want to copy the data into,
// 	// and the number of arguments must be exactly the same as the number of
// 	// columns returned by your statement.
//  // Convert sql.ErrNoRows to ErrNoRecord so upper layers depend on model/domain errors, not database-specific errors.
// 	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return Snippet{}, ErrNoRecord
// 		} else {
// 			return Snippet{}, err
// 		}
// 	}

//		return s, nil
//	}
//
// This will return a specific snippet based on its ID.
func (m *SnippetModel) Get(id int) (Snippet, error) {
	var s Snippet

	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() AND id = ?`
	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		} else {
			return Snippet{}, err
		}
	}

	return s, nil
}

// This will return the 10 most recently created snippets that haven't expired.
func (m *SnippetModel) Latest() ([]Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() ORDER BY created DESC LIMIT 10`

	// Use the Query() method on the connection pool to execute our
	// SQL statement. This returns a sql.Rows resultset containing the result of
	// our query.
	rows, err := m.DB.Query(stmt)

	if err != nil {
		return nil, err
	}

	// We defer rows.Close() to ensure the sql.Rows resultset is
	// always properly closed before the Latest() method returns. This defer
	// statement should come *after* you check for an error from the Query()
	// method. Otherwise, if Query() returns an error, you'll get a panic
	// trying to close a nil resultset.
	// Important: Closing a resultset with defer rows.Close() is critical in the code above.
	// As long as a resultset is open it will keep the underlying database connection open…
	// so if something goes wrong in this method and the resultset isn’t closed, it can rapidly
	// lead to all the connections in your pool being used up.
	defer rows.Close()

	var snippets []Snippet

	for rows.Next() {
		var s Snippet
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}
	// When the rows.Next() loop has finished we call rows.Err() to retrieve any
	// error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
