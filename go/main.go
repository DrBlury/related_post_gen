package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const topN = 5
const InitialTagMapSize = 100

type isize uint16

type Post struct {
	ID    string   `json:"_id"`
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

type RelatedPosts struct {
	ID      string      `json:"_id"`
	Tags    *[]string   `json:"tags"`
	Related [topN]*Post `json:"related"`
}

type PostWithSharedTags struct {
	Post       isize
	SharedTags isize
}

func buildTagMap(posts []Post) map[string][]isize {
	tagMap := make(map[string][]isize, InitialTagMapSize)
	for i, post := range posts {
		for _, tag := range post.Tags {
			tagMap[tag] = append(tagMap[tag], isize(i))
		}
	}
	return tagMap
}

func computeRelatedPosts(posts []Post, tagMap map[string][]isize) []RelatedPosts {
	postsLen := len(posts)
	allRelatedPosts := make([]RelatedPosts, postsLen)
	taggedPostCount := make([]isize, postsLen)

	for i := range posts {
		for j := range taggedPostCount {
			taggedPostCount[j] = 0
		}
		// Count the number of tags shared between posts
		for _, tag := range posts[i].Tags {
			for _, otherPostIdx := range tagMap[tag] {
				taggedPostCount[otherPostIdx]++
			}
		}
		taggedPostCount[i] = 0 // Don't count self
		top5 := [topN]PostWithSharedTags{}
		minTags := isize(0)

		for j, count := range taggedPostCount {
			if count > minTags {
				// Find the position to insert
				pos := 4
				for pos >= 0 && top5[pos].SharedTags < count {
					pos--
				}
				pos++

				// Shift and insert
				if pos < 4 {
					copy(top5[pos+1:], top5[pos:4])
				}

				top5[pos] = PostWithSharedTags{Post: isize(j), SharedTags: count}
				minTags = top5[4].SharedTags
			}
		}
		// Convert indexes back to Post pointers
		topPosts := [topN]*Post{}
		for idx, t := range top5 {
			topPosts[idx] = &posts[t.Post]
		}

		allRelatedPosts[i] = RelatedPosts{
			ID:      posts[i].ID,
			Tags:    &posts[i].Tags,
			Related: topPosts,
		}
	}
	return allRelatedPosts
}

func readPosts() []Post {
	file, _ := os.Open("../posts.json")
	var posts = []Post{}
	err := json.NewDecoder(file).Decode(&posts)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return posts
}

func main() {
	posts := readPosts()
	start := time.Now()
	tagMap := buildTagMap(posts)
	allRelatedPosts := computeRelatedPosts(posts, tagMap)

	fmt.Println("Processing time (w/o IO):", time.Since(start))
	file, _ := os.Create("../related_posts_go.json")
	_ = json.NewEncoder(file).Encode(allRelatedPosts)
}
