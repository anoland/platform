// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package api

import (
	"github.com/mattermost/platform/model"
	"github.com/mattermost/platform/utils"
	"net/http"
	//"strings"
	"fmt"
	"testing"
	"time"
)

func TestCreatePost(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	team := th.BasicTeam
	team2 := th.CreateTeam(th.BasicClient)
	user1 := th.BasicUser
	user3 := th.CreateUser(th.BasicClient)
	LinkUserToTeam(user3, team2)
	channel1 := th.BasicChannel
	channel2 := th.CreateChannel(Client, team)

	filenames := []string{"/12345678901234567890123456/12345678901234567890123456/12345678901234567890123456/test.png", "/" + channel1.Id + "/" + user1.Id + "/test.png", "www.mattermost.com/fake/url", "junk"}

	post1 := &model.Post{ChannelId: channel1.Id, Message: "#hashtag a" + model.NewId() + "a", Filenames: filenames}
	rpost1, err := Client.CreatePost(post1)
	if err != nil {
		t.Fatal(err)
	}

	if rpost1.Data.(*model.Post).Message != post1.Message {
		t.Fatal("message didn't match")
	}

	if rpost1.Data.(*model.Post).Hashtags != "#hashtag" {
		t.Fatal("hashtag didn't match")
	}

	if len(rpost1.Data.(*model.Post).Filenames) != 2 {
		t.Fatal("filenames didn't parse correctly")
	}

	post2 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: rpost1.Data.(*model.Post).Id}
	rpost2, err := Client.CreatePost(post2)
	if err != nil {
		t.Fatal(err)
	}

	post3 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: rpost1.Data.(*model.Post).Id, ParentId: rpost2.Data.(*model.Post).Id}
	_, err = Client.CreatePost(post3)
	if err != nil {
		t.Fatal(err)
	}

	post4 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: "junk"}
	_, err = Client.CreatePost(post4)
	if err.StatusCode != http.StatusBadRequest {
		t.Fatal("Should have been invalid param")
	}

	post5 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: rpost1.Data.(*model.Post).Id, ParentId: "junk"}
	_, err = Client.CreatePost(post5)
	if err.StatusCode != http.StatusBadRequest {
		t.Fatal("Should have been invalid param")
	}

	post1c2 := &model.Post{ChannelId: channel2.Id, Message: "a" + model.NewId() + "a"}
	rpost1c2, err := Client.CreatePost(post1c2)

	post2c2 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: rpost1c2.Data.(*model.Post).Id}
	_, err = Client.CreatePost(post2c2)
	if err.StatusCode != http.StatusBadRequest {
		t.Fatal("Should have been invalid param")
	}

	post6 := &model.Post{ChannelId: "junk", Message: "a" + model.NewId() + "a"}
	_, err = Client.CreatePost(post6)
	if err.StatusCode != http.StatusForbidden {
		t.Fatal("Should have been forbidden")
	}

	th.LoginBasic2()

	post7 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	_, err = Client.CreatePost(post7)
	if err.StatusCode != http.StatusForbidden {
		t.Fatal("Should have been forbidden")
	}

	Client.LoginByEmail(team2.Name, user3.Email, user3.Password)
	Client.SetTeamId(team2.Id)
	channel3 := th.CreateChannel(Client, team2)

	post8 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	_, err = Client.CreatePost(post8)
	if err.StatusCode != http.StatusForbidden {
		t.Fatal("Should have been forbidden")
	}

	if _, err = Client.DoApiPost("/channels/"+channel3.Id+"/create", "garbage"); err == nil {
		t.Fatal("should have been an error")
	}
}

func TestUpdatePost(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel

	post1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	rpost1, err := Client.CreatePost(post1)
	if err != nil {
		t.Fatal(err)
	}

	if rpost1.Data.(*model.Post).Message != post1.Message {
		t.Fatal("full name didn't match")
	}

	post2 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: rpost1.Data.(*model.Post).Id}
	rpost2, err := Client.CreatePost(post2)
	if err != nil {
		t.Fatal(err)
	}

	msg2 := "a" + model.NewId() + " update post 1"
	rpost2.Data.(*model.Post).Message = msg2
	if rupost2, err := Client.UpdatePost(rpost2.Data.(*model.Post)); err != nil {
		t.Fatal(err)
	} else {
		if rupost2.Data.(*model.Post).Message != msg2 {
			t.Fatal("failed to updates")
		}
	}

	msg1 := "#hashtag a" + model.NewId() + " update post 2"
	rpost1.Data.(*model.Post).Message = msg1
	if rupost1, err := Client.UpdatePost(rpost1.Data.(*model.Post)); err != nil {
		t.Fatal(err)
	} else {
		if rupost1.Data.(*model.Post).Message != msg1 && rupost1.Data.(*model.Post).Hashtags != "#hashtag" {
			t.Fatal("failed to updates")
		}
	}

	up12 := &model.Post{Id: rpost1.Data.(*model.Post).Id, ChannelId: channel1.Id, Message: "a" + model.NewId() + " updaet post 1 update 2"}
	if rup12, err := Client.UpdatePost(up12); err != nil {
		t.Fatal(err)
	} else {
		if rup12.Data.(*model.Post).Message != up12.Message {
			t.Fatal("failed to updates")
		}
	}
}

func TestGetPosts(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel

	time.Sleep(10 * time.Millisecond)
	post1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post1 = Client.Must(Client.CreatePost(post1)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post1a1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: post1.Id}
	post1a1 = Client.Must(Client.CreatePost(post1a1)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post2 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post2 = Client.Must(Client.CreatePost(post2)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post3 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post3 = Client.Must(Client.CreatePost(post3)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post3a1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: post3.Id}
	post3a1 = Client.Must(Client.CreatePost(post3a1)).Data.(*model.Post)

	r1 := Client.Must(Client.GetPosts(channel1.Id, 0, 2, "")).Data.(*model.PostList)

	if r1.Order[0] != post3a1.Id {
		t.Fatal("wrong order")
	}

	if r1.Order[1] != post3.Id {
		t.Fatal("wrong order")
	}

	if len(r1.Posts) != 2 { // 3a1 and 3; 3a1's parent already there
		t.Fatal("wrong size")
	}

	r2 := Client.Must(Client.GetPosts(channel1.Id, 2, 2, "")).Data.(*model.PostList)

	if r2.Order[0] != post2.Id {
		t.Fatal("wrong order")
	}

	if r2.Order[1] != post1a1.Id {
		t.Fatal("wrong order")
	}

	if len(r2.Posts) != 3 { // 2 and 1a1; + 1a1's parent
		t.Log(r2.Posts)
		t.Fatal("wrong size")
	}
}

func TestGetPostsSince(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel

	time.Sleep(10 * time.Millisecond)
	post0 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post0 = Client.Must(Client.CreatePost(post0)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post1 = Client.Must(Client.CreatePost(post1)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post1a1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: post1.Id}
	post1a1 = Client.Must(Client.CreatePost(post1a1)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post2 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post2 = Client.Must(Client.CreatePost(post2)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post3 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post3 = Client.Must(Client.CreatePost(post3)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post3a1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: post3.Id}
	post3a1 = Client.Must(Client.CreatePost(post3a1)).Data.(*model.Post)

	r1 := Client.Must(Client.GetPostsSince(channel1.Id, post1.CreateAt)).Data.(*model.PostList)

	if r1.Order[0] != post3a1.Id {
		t.Fatal("wrong order")
	}

	if r1.Order[1] != post3.Id {
		t.Fatal("wrong order")
	}

	if len(r1.Posts) != 5 {
		t.Fatal("wrong size")
	}

	now := model.GetMillis()
	r2 := Client.Must(Client.GetPostsSince(channel1.Id, now)).Data.(*model.PostList)

	if len(r2.Posts) != 0 {
		t.Fatal("should have been empty")
	}

	post2.Message = "new message"
	Client.Must(Client.UpdatePost(post2))

	r3 := Client.Must(Client.GetPostsSince(channel1.Id, now)).Data.(*model.PostList)

	if len(r3.Order) != 2 { // 2 because deleted post is returned as well
		t.Fatal("missing post update")
	}
}

func TestGetPostsBeforeAfter(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel

	time.Sleep(10 * time.Millisecond)
	post0 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post0 = Client.Must(Client.CreatePost(post0)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post1 = Client.Must(Client.CreatePost(post1)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post1a1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: post1.Id}
	post1a1 = Client.Must(Client.CreatePost(post1a1)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post2 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post2 = Client.Must(Client.CreatePost(post2)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post3 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post3 = Client.Must(Client.CreatePost(post3)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post3a1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: post3.Id}
	post3a1 = Client.Must(Client.CreatePost(post3a1)).Data.(*model.Post)

	r1 := Client.Must(Client.GetPostsBefore(channel1.Id, post1a1.Id, 0, 10, "")).Data.(*model.PostList)

	if r1.Order[0] != post1.Id {
		t.Fatal("wrong order")
	}

	if r1.Order[1] != post0.Id {
		t.Fatal("wrong order")
	}

	if len(r1.Posts) != 3 {
		t.Log(r1.Posts)
		t.Fatal("wrong size")
	}

	r2 := Client.Must(Client.GetPostsAfter(channel1.Id, post3a1.Id, 0, 3, "")).Data.(*model.PostList)

	if len(r2.Posts) != 0 {
		t.Fatal("should have been empty")
	}

	post2.Message = "new message"
	Client.Must(Client.UpdatePost(post2))

	r3 := Client.Must(Client.GetPostsAfter(channel1.Id, post1a1.Id, 0, 2, "")).Data.(*model.PostList)

	if r3.Order[0] != post3.Id {
		t.Fatal("wrong order")
	}

	if r3.Order[1] != post2.Id {
		t.Fatal("wrong order")
	}

	if len(r3.Order) != 2 {
		t.Fatal("missing post update")
	}
}

func TestSearchPosts(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel

	post1 := &model.Post{ChannelId: channel1.Id, Message: "search for post1"}
	post1 = Client.Must(Client.CreatePost(post1)).Data.(*model.Post)

	post2 := &model.Post{ChannelId: channel1.Id, Message: "search for post2"}
	post2 = Client.Must(Client.CreatePost(post2)).Data.(*model.Post)

	post3 := &model.Post{ChannelId: channel1.Id, Message: "#hashtag search for post3"}
	post3 = Client.Must(Client.CreatePost(post3)).Data.(*model.Post)

	post4 := &model.Post{ChannelId: channel1.Id, Message: "hashtag for post4"}
	post4 = Client.Must(Client.CreatePost(post4)).Data.(*model.Post)

	r1 := Client.Must(Client.SearchPosts("search")).Data.(*model.PostList)

	if len(r1.Order) != 3 {
		t.Fatal("wrong serach")
	}

	r2 := Client.Must(Client.SearchPosts("post2")).Data.(*model.PostList)

	if len(r2.Order) != 1 && r2.Order[0] == post2.Id {
		t.Fatal("wrong serach")
	}

	r3 := Client.Must(Client.SearchPosts("#hashtag")).Data.(*model.PostList)

	if len(r3.Order) != 1 && r3.Order[0] == post3.Id {
		t.Fatal("wrong serach")
	}

	if r4 := Client.Must(Client.SearchPosts("*")).Data.(*model.PostList); len(r4.Order) != 0 {
		t.Fatal("searching for just * shouldn't return any results")
	}
}

func TestSearchHashtagPosts(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel

	post1 := &model.Post{ChannelId: channel1.Id, Message: "#sgtitlereview with space"}
	post1 = Client.Must(Client.CreatePost(post1)).Data.(*model.Post)

	post2 := &model.Post{ChannelId: channel1.Id, Message: "#sgtitlereview\n with return"}
	post2 = Client.Must(Client.CreatePost(post2)).Data.(*model.Post)

	post3 := &model.Post{ChannelId: channel1.Id, Message: "no hashtag"}
	post3 = Client.Must(Client.CreatePost(post3)).Data.(*model.Post)

	r1 := Client.Must(Client.SearchPosts("#sgtitlereview")).Data.(*model.PostList)

	if len(r1.Order) != 2 {
		t.Fatal("wrong search")
	}
}

func TestSearchPostsInChannel(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel
	team := th.BasicTeam

	post1 := &model.Post{ChannelId: channel1.Id, Message: "sgtitlereview with space"}
	post1 = Client.Must(Client.CreatePost(post1)).Data.(*model.Post)

	channel2 := &model.Channel{DisplayName: "TestGetPosts", Name: "a" + model.NewId() + "a", Type: model.CHANNEL_OPEN, TeamId: team.Id}
	channel2 = Client.Must(Client.CreateChannel(channel2)).Data.(*model.Channel)

	channel3 := &model.Channel{DisplayName: "TestGetPosts", Name: "a" + model.NewId() + "a", Type: model.CHANNEL_OPEN, TeamId: team.Id}
	channel3 = Client.Must(Client.CreateChannel(channel3)).Data.(*model.Channel)

	post2 := &model.Post{ChannelId: channel2.Id, Message: "sgtitlereview\n with return"}
	post2 = Client.Must(Client.CreatePost(post2)).Data.(*model.Post)

	post3 := &model.Post{ChannelId: channel2.Id, Message: "other message with no return"}
	post3 = Client.Must(Client.CreatePost(post3)).Data.(*model.Post)

	post4 := &model.Post{ChannelId: channel3.Id, Message: "other message with no return"}
	post4 = Client.Must(Client.CreatePost(post4)).Data.(*model.Post)

	if result := Client.Must(Client.SearchPosts("channel:")).Data.(*model.PostList); len(result.Order) != 0 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("in:")).Data.(*model.PostList); len(result.Order) != 0 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("channel:" + channel1.Name)).Data.(*model.PostList); len(result.Order) != 2 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("in: " + channel2.Name)).Data.(*model.PostList); len(result.Order) != 2 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("channel: " + channel2.Name)).Data.(*model.PostList); len(result.Order) != 2 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("ChAnNeL: " + channel2.Name)).Data.(*model.PostList); len(result.Order) != 2 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("sgtitlereview")).Data.(*model.PostList); len(result.Order) != 2 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("sgtitlereview channel:" + channel1.Name)).Data.(*model.PostList); len(result.Order) != 1 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("sgtitlereview in: " + channel2.Name)).Data.(*model.PostList); len(result.Order) != 1 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("sgtitlereview channel: " + channel2.Name)).Data.(*model.PostList); len(result.Order) != 1 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("channel: " + channel2.Name + " channel: " + channel3.Name)).Data.(*model.PostList); len(result.Order) != 3 {
		t.Fatalf("wrong number of posts returned :) %v :) %v", result.Posts, result.Order)
	}
}

func TestSearchPostsFromUser(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel
	team := th.BasicTeam
	user1 := th.BasicUser
	user2 := th.BasicUser2
	channel2 := th.CreateChannel(Client, team)
	Client.Must(Client.AddChannelMember(channel1.Id, th.BasicUser2.Id))
	Client.Must(Client.AddChannelMember(channel2.Id, th.BasicUser2.Id))
	user3 := th.CreateUser(Client)
	LinkUserToTeam(user3, team)
	Client.Must(Client.AddChannelMember(channel1.Id, user3.Id))
	Client.Must(Client.AddChannelMember(channel2.Id, user3.Id))

	post1 := &model.Post{ChannelId: channel1.Id, Message: "sgtitlereview with space"}
	post1 = Client.Must(Client.CreatePost(post1)).Data.(*model.Post)

	th.LoginBasic2()

	post2 := &model.Post{ChannelId: channel2.Id, Message: "sgtitlereview\n with return"}
	post2 = Client.Must(Client.CreatePost(post2)).Data.(*model.Post)

	if result := Client.Must(Client.SearchPosts("from: " + user1.Username)).Data.(*model.PostList); len(result.Order) != 2 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("from: " + user2.Username)).Data.(*model.PostList); len(result.Order) != 1 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("from: " + user2.Username + " sgtitlereview")).Data.(*model.PostList); len(result.Order) != 1 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	post3 := &model.Post{ChannelId: channel1.Id, Message: "hullo"}
	post3 = Client.Must(Client.CreatePost(post3)).Data.(*model.Post)

	if result := Client.Must(Client.SearchPosts("from: " + user2.Username + " in:" + channel1.Name)).Data.(*model.PostList); len(result.Order) != 1 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	Client.LoginByEmail(team.Name, user3.Email, user3.Password)

	// wait for the join/leave messages to be created for user3 since they're done asynchronously
	time.Sleep(100 * time.Millisecond)

	if result := Client.Must(Client.SearchPosts("from: " + user2.Username)).Data.(*model.PostList); len(result.Order) != 2 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("from: " + user2.Username + " from: " + user3.Username)).Data.(*model.PostList); len(result.Order) != 2 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	if result := Client.Must(Client.SearchPosts("from: " + user2.Username + " from: " + user3.Username + " in:" + channel2.Name)).Data.(*model.PostList); len(result.Order) != 1 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}

	post4 := &model.Post{ChannelId: channel2.Id, Message: "coconut"}
	post4 = Client.Must(Client.CreatePost(post4)).Data.(*model.Post)

	if result := Client.Must(Client.SearchPosts("from: " + user2.Username + " from: " + user3.Username + " in:" + channel2.Name + " coconut")).Data.(*model.PostList); len(result.Order) != 1 {
		t.Fatalf("wrong number of posts returned %v", len(result.Order))
	}
}

func TestGetPostsCache(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel

	time.Sleep(10 * time.Millisecond)
	post1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post1 = Client.Must(Client.CreatePost(post1)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post2 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post2 = Client.Must(Client.CreatePost(post2)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post3 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post3 = Client.Must(Client.CreatePost(post3)).Data.(*model.Post)

	etag := Client.Must(Client.GetPosts(channel1.Id, 0, 2, "")).Etag

	// test etag caching
	if cache_result, err := Client.GetPosts(channel1.Id, 0, 2, etag); err != nil {
		t.Fatal(err)
	} else if cache_result.Data.(*model.PostList) != nil {
		t.Log(cache_result.Data)
		t.Fatal("cache should be empty")
	}

	etag = Client.Must(Client.GetPost(channel1.Id, post1.Id, "")).Etag

	// test etag caching
	if cache_result, err := Client.GetPost(channel1.Id, post1.Id, etag); err != nil {
		t.Fatal(err)
	} else if cache_result.Data.(*model.PostList) != nil {
		t.Log(cache_result.Data)
		t.Fatal("cache should be empty")
	}

}

func TestDeletePosts(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel
	UpdateUserToTeamAdmin(th.BasicUser2, th.BasicTeam)

	time.Sleep(10 * time.Millisecond)
	post1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post1 = Client.Must(Client.CreatePost(post1)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post1a1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: post1.Id}
	post1a1 = Client.Must(Client.CreatePost(post1a1)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post1a2 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: post1.Id, ParentId: post1a1.Id}
	post1a2 = Client.Must(Client.CreatePost(post1a2)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post2 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post2 = Client.Must(Client.CreatePost(post2)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post3 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post3 = Client.Must(Client.CreatePost(post3)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	post3a1 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a", RootId: post3.Id}
	post3a1 = Client.Must(Client.CreatePost(post3a1)).Data.(*model.Post)

	time.Sleep(10 * time.Millisecond)
	Client.Must(Client.DeletePost(channel1.Id, post3.Id))

	r2 := Client.Must(Client.GetPosts(channel1.Id, 0, 10, "")).Data.(*model.PostList)

	if len(r2.Posts) != 5 {
		t.Fatal("should have returned 4 items")
	}

	time.Sleep(10 * time.Millisecond)
	post4 := &model.Post{ChannelId: channel1.Id, Message: "a" + model.NewId() + "a"}
	post4 = Client.Must(Client.CreatePost(post4)).Data.(*model.Post)

	th.LoginBasic2()

	Client.Must(Client.DeletePost(channel1.Id, post4.Id))
}

func TestEmailMention(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel

	post1 := &model.Post{ChannelId: channel1.Id, Message: th.BasicUser.Username}
	post1 = Client.Must(Client.CreatePost(post1)).Data.(*model.Post)

	// No easy way to verify the email was sent, but this will at least cause the server to throw errors if the code is broken

}

func TestFuzzyPosts(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel

	filenames := []string{"junk"}

	for i := 0; i < len(utils.FUZZY_STRINGS_POSTS); i++ {
		post := &model.Post{ChannelId: channel1.Id, Message: utils.FUZZY_STRINGS_POSTS[i], Filenames: filenames}

		_, err := Client.CreatePost(post)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TODO XXX FIX ME - Need to figure out how DM work with users as FCO

// func TestMakeDirectChannelVisible(t *testing.T) {
// 	Setup()

// 	team := &model.Team{DisplayName: "Name", Name: "z-z-" + model.NewId() + "a", Email: "test@nowhere.com", Type: model.TEAM_OPEN}
// 	team = Client.Must(Client.CreateTeam(team)).Data.(*model.Team)

// 	user1 := &model.User{Email: model.NewId() + "success+test@simulator.amazonses.com", Nickname: "Corey Hulen", Password: "pwd"}
// 	user1 = Client.Must(Client.CreateUser(user1, "")).Data.(*model.User)
// 	LinkUserToTeam(user1.Id, team.Id)
// 	store.Must(Srv.Store.User().VerifyEmail(user1.Id))

// 	user2 := &model.User{Email: model.NewId() + "success+test@simulator.amazonses.com", Nickname: "Corey Hulen", Password: "pwd"}
// 	user2 = Client.Must(Client.CreateUser(user2, "")).Data.(*model.User)
// 	LinkUserToTeam(user2.Id, team.Id)
// 	store.Must(Srv.Store.User().VerifyEmail(user2.Id))

// 	// user2 will be created with prefs created to show user1 in the sidebar so set that to false to get rid of it
// 	Client.LoginByEmail(team.Name, user2.Email, "pwd")

// 	preferences := &model.Preferences{
// 		{
// 			UserId:   user2.Id,
// 			Category: model.PREFERENCE_CATEGORY_DIRECT_CHANNEL_SHOW,
// 			Name:     user1.Id,
// 			Value:    "false",
// 		},
// 	}
// 	Client.Must(Client.SetPreferences(preferences))

// 	Client.LoginByEmail(team.Name, user1.Email, "pwd")

// 	channel := Client.Must(Client.CreateDirectChannel(map[string]string{"user_id": user2.Id})).Data.(*model.Channel)

// 	makeDirectChannelVisible(team.Id, channel.Id)

// 	if result, err := Client.GetPreference(model.PREFERENCE_CATEGORY_DIRECT_CHANNEL_SHOW, user2.Id); err != nil {
// 		t.Fatal("Errored trying to set direct channel to be visible for user1")
// 	} else if pref := result.Data.(*model.Preference); pref.Value != "true" {
// 		t.Fatal("Failed to set direct channel to be visible for user1")
// 	}

// 	Client.LoginByEmail(team.Name, user2.Email, "pwd")

// 	if result, err := Client.GetPreference(model.PREFERENCE_CATEGORY_DIRECT_CHANNEL_SHOW, user1.Id); err != nil {
// 		t.Fatal("Errored trying to set direct channel to be visible for user2")
// 	} else if pref := result.Data.(*model.Preference); pref.Value != "true" {
// 		t.Fatal("Failed to set direct channel to be visible for user2")
// 	}
// }

func TestGetOutOfChannelMentions(t *testing.T) {
	th := Setup().InitBasic()
	Client := th.BasicClient
	channel1 := th.BasicChannel
	team1 := th.BasicTeam
	user1 := th.BasicUser
	user2 := th.BasicUser2
	user3 := th.CreateUser(Client)
	LinkUserToTeam(user3, team1)

	var allProfiles map[string]*model.User
	if result := <-Srv.Store.User().GetProfiles(team1.Id); result.Err != nil {
		t.Fatal(result.Err)
	} else {
		allProfiles = result.Data.(map[string]*model.User)
	}

	var members []model.ChannelMember
	if result := <-Srv.Store.Channel().GetMembers(channel1.Id); result.Err != nil {
		t.Fatal(result.Err)
	} else {
		members = result.Data.([]model.ChannelMember)
	}

	// test a post that doesn't @mention anybody
	post1 := &model.Post{ChannelId: channel1.Id, Message: fmt.Sprintf("%v %v %v", user1.Username, user2.Username, user3.Username)}
	if mentioned := getOutOfChannelMentions(post1, allProfiles, members); len(mentioned) != 0 {
		t.Fatalf("getOutOfChannelMentions returned %v when no users were mentioned", mentioned)
	}

	// test a post that @mentions someone in the channel
	post2 := &model.Post{ChannelId: channel1.Id, Message: fmt.Sprintf("@%v is %v", user1.Username, user1.Username)}
	if mentioned := getOutOfChannelMentions(post2, allProfiles, members); len(mentioned) != 0 {
		t.Fatalf("getOutOfChannelMentions returned %v when only users in the channel were mentioned", mentioned)
	}

	// test a post that @mentions someone not in the channel
	post3 := &model.Post{ChannelId: channel1.Id, Message: fmt.Sprintf("@%v and @%v aren't in the channel", user2.Username, user3.Username)}
	if mentioned := getOutOfChannelMentions(post3, allProfiles, members); len(mentioned) != 2 || (mentioned[0].Id != user2.Id && mentioned[0].Id != user3.Id) || (mentioned[1].Id != user2.Id && mentioned[1].Id != user3.Id) {
		t.Fatalf("getOutOfChannelMentions returned %v when two users outside the channel were mentioned", mentioned)
	}

	// test a post that @mentions someone not in the channel as well as someone in the channel
	post4 := &model.Post{ChannelId: channel1.Id, Message: fmt.Sprintf("@%v and @%v might be in the channel", user2.Username, user1.Username)}
	if mentioned := getOutOfChannelMentions(post4, allProfiles, members); len(mentioned) != 1 || mentioned[0].Id != user2.Id {
		t.Fatalf("getOutOfChannelMentions returned %v when someone in the channel and someone  outside the channel were mentioned", mentioned)
	}

	Client.Must(Client.Logout())

	team2 := th.CreateTeam(Client)
	user4 := th.CreateUser(Client)
	LinkUserToTeam(user4, team2)

	Client.Must(Client.LoginByEmail(team2.Name, user4.Email, user4.Password))
	Client.SetTeamId(team2.Id)

	channel2 := &model.Channel{DisplayName: "Test API Name", Name: "a" + model.NewId() + "a", Type: model.CHANNEL_OPEN, TeamId: team2.Id}
	channel2 = Client.Must(Client.CreateChannel(channel2)).Data.(*model.Channel)

	if result := <-Srv.Store.User().GetProfiles(team2.Id); result.Err != nil {
		t.Fatal(result.Err)
	} else {
		allProfiles = result.Data.(map[string]*model.User)
	}

	if result := <-Srv.Store.Channel().GetMembers(channel2.Id); result.Err != nil {
		t.Fatal(result.Err)
	} else {
		members = result.Data.([]model.ChannelMember)
	}

	// test a post that @mentions someone on a different team
	post5 := &model.Post{ChannelId: channel2.Id, Message: fmt.Sprintf("@%v and @%v might be in the channel", user2.Username, user3.Username)}
	if mentioned := getOutOfChannelMentions(post5, allProfiles, members); len(mentioned) != 0 {
		t.Fatalf("getOutOfChannelMentions returned %v when two users on a different team were mentioned", mentioned)
	}
}
