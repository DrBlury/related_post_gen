package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	topN              = 5
	InitialTagMapSize = 100
	INPUT_FILE        = "../posts.json"
	OUTPUT_FILE       = "../related_posts_go.json"
)

type isize uint16

type Post struct {
	ID    string    `json:"_id"`
	Title string    `json:"title"`
	Tags  *[]string `json:"tags"`
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
		for _, tag := range *post.Tags {
			tagMap[tag] = append(tagMap[tag], isize(i))
		}
	}
	return tagMap
}

func computeRelatedPosts(posts []Post, tagMap map[string][]isize) []RelatedPosts {
	allRelatedPosts := make([]RelatedPosts, 0, len(posts))
	taggedPostCount := make([]isize, len(posts))

	for postIdx := range posts {
		// Reset the tag count for each post
		for j := range taggedPostCount {
			taggedPostCount[j] = 0
		}

		// Count the number of tags shared between posts
		for _, tag := range *posts[postIdx].Tags {
			for _, otherPostIdx := range tagMap[tag] {
				taggedPostCount[otherPostIdx]++
			}
		}
		top5 := [topN]PostWithSharedTags{}
		taggedPostCount[postIdx] = 0 // Don't count self
		for j, count := range taggedPostCount {
			if count > top5[0].SharedTags {
				top5[0] = PostWithSharedTags{Post: isize(j), SharedTags: count}
				for k := 0; k < topN-1; k++ {
					if top5[k].SharedTags > top5[k+1].SharedTags {
						top5[k], top5[k+1] = top5[k+1], top5[k]
					}
				}
			}
		}

		topPosts := [topN]*Post{}
		for idx, t := range top5 {
			topPosts[idx] = &posts[t.Post]
		}

		allRelatedPosts = append(allRelatedPosts, RelatedPosts{
			ID:      posts[postIdx].ID,
			Tags:    posts[postIdx].Tags,
			Related: topPosts,
		})
	}

	return allRelatedPosts
}

func readPosts() []Post {
	file, _ := os.Open(INPUT_FILE)
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
	file, _ := os.Create(OUTPUT_FILE)
	_ = json.NewEncoder(file).Encode(allRelatedPosts)
}
