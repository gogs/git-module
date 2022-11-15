package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepoGrepSimple(t *testing.T) {
	pattern := "programmingPoints"
	expect := []GrepResult{
		{
			TreeID: "HEAD", Path: "src/Main.groovy", Line: 7, Column: 5, Text: "int programmingPoints = 10",
		},
		{
			TreeID: "HEAD", Path: "src/Main.groovy", Line: 10, Column: 33, Text: `println "${name} has at least ${programmingPoints} programming points."`,
		},
		{
			TreeID: "HEAD", Path: "src/Main.groovy", Line: 11, Column: 12, Text: `println "${programmingPoints} squared is ${square(programmingPoints)}"`,
		},
		{
			TreeID: "HEAD", Path: "src/Main.groovy", Line: 12, Column: 12, Text: `println "${programmingPoints} divided by 2 bonus points is ${divide(programmingPoints, 2)}"`,
		},
		{
			TreeID: "HEAD", Path: "src/Main.groovy", Line: 13, Column: 12, Text: `println "${programmingPoints} minus 7 bonus points is ${subtract(programmingPoints, 7)}"`,
		},
		{
			TreeID: "HEAD", Path: "src/Main.groovy", Line: 14, Column: 12, Text: `println "${programmingPoints} plus 3 bonus points is ${sum(programmingPoints, 3)}"`,
		},
	}
	results, err := testrepo.Grep(pattern)
	assert.NoError(t, err)
	for i, result := range results {
		assert.Equal(t, expect[i], *result)
	}
}

func TestRepoGrepIgnoreCase(t *testing.T) {
	pattern := "Hello"
	expect := []GrepResult{
		{
			TreeID: "HEAD", Path: "README.txt", Line: 9, Column: 36, Text: "* git@github.com:matthewmccullough/hellogitworld.git",
		},
		{
			TreeID: "HEAD", Path: "README.txt", Line: 10, Column: 38, Text: "* git://github.com/matthewmccullough/hellogitworld.git",
		},
		{
			TreeID: "HEAD", Path: "README.txt", Line: 11, Column: 58, Text: "* https://matthewmccullough@github.com/matthewmccullough/hellogitworld.git",
		},
		{
			TreeID: "HEAD", Path: "src/Main.groovy", Line: 9, Column: 10, Text: `println "Hello ${name}"`,
		},
		{
			TreeID: "HEAD", Path: "src/main/java/com/github/App.java", Line: 4, Column: 4, Text: " * Hello again",
		},
		{
			TreeID: "HEAD", Path: "src/main/java/com/github/App.java", Line: 5, Column: 4, Text: " * Hello world!",
		},
		{
			TreeID: "HEAD", Path: "src/main/java/com/github/App.java", Line: 6, Column: 4, Text: " * Hello",
		},
		{
			TreeID: "HEAD", Path: "src/main/java/com/github/App.java", Line: 13, Column: 30, Text: `        System.out.println( "Hello World!" );`,
		},
	}
	results, err := testrepo.Grep(pattern, GrepOptions{IgnoreCase: true})
	assert.NoError(t, err)
	for i, result := range results {
		assert.Equal(t, expect[i], *result)
	}
}

func TestRepoGrepRegex(t *testing.T) {
	pattern := "Hello\\sW\\w+"
	expect := []GrepResult{
		{
			TreeID: "HEAD", Path: "src/main/java/com/github/App.java", Line: 13, Column: 30, Text: `        System.out.println( "Hello World!" );`,
		},
	}
	results, err := testrepo.Grep(pattern, GrepOptions{ExtendedRegex: true})
	assert.NoError(t, err)
	for i, result := range results {
		assert.Equal(t, expect[i], *result)
	}
}

func TestRepoGrepWord(t *testing.T) {
	pattern := "Hello\\sW\\w+"
	expect := []GrepResult{
		{
			TreeID: "HEAD", Path: "src/main/java/com/github/App.java", Line: 13, Column: 36, Text: `        System.out.println( "Hello World!" );`,
		},
	}
	results, err := testrepo.Grep(pattern, GrepOptions{WordMatch: true})
	assert.NoError(t, err)
	for i, result := range results {
		assert.Equal(t, expect[i], *result)
	}
}
