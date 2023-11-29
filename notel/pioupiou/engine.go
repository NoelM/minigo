package pioupiou

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/cockroachdb/pebble"
)

var infoLog = log.New(os.Stdout, "[pioupiou] info:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var warnLog = log.New(os.Stdout, "[pioupiou] warn:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var errorLog = log.New(os.Stdout, "[pioupiou] error:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

const (
	FilCmdStr        = "/FIL"
	MessageCmdStr    = "/MSG"
	NotifCmdStr      = "/NOT"
	RechercheCmdStr  = "/CRC"
	ProfilCmdStr     = "/PRO"
	AbonnementCmdStr = "/ABO"
	EditerCmdStr     = "/EDI"
	AnnuaireCmdStr   = "/ANU"
)

const (
	FilId = iota
	MessageId
	NotifId
	RechercheId
	ProfilId
	AbonnementId
	EditerId
	AnnuaireId
)

var CuicuiCommands = map[string]int{
	FilCmdStr:        FilId,
	MessageCmdStr:    MessageId,
	NotifCmdStr:      NotifId,
	RechercheCmdStr:  RechercheId,
	ProfilCmdStr:     ProfilId,
	AbonnementCmdStr: AbonnementId,
	EditerCmdStr:     EditerId,
	AnnuaireCmdStr:   AnnuaireId,
}

func NewServiceCuicui() int {
	return 0
}

func ParseCommandString(cmd string) int {
	for cmdStr, cmdId := range CuicuiCommands {
		if cmd == cmdStr {
			return cmdId
		}
	}
	return -1
}

func PrintErrorMessage(msg string) error {
	return nil
}

var PiouPiou, _ = NewPPEngine("/media/core/")

type PPEngine struct {
	rootDir      string
	usersDB      *pebble.DB
	followsDB    *pebble.DB
	activitiesDB *pebble.DB
	postsDB      *pebble.DB
	minPostFeed  map[string]int
}

const PPUsersDir = "pp_users_db"

type PPUser struct {
	Pseudo      string    `json:"pseudo"`
	Bio         string    `json:"bio"`
	Tel         string    `json:"tel"`
	Location    string    `json:"location"`
	LastConnect time.Time `json:"last_connect"`
}

const PPFollowsDir = "pp_follows_db"

type PPFollow struct {
	Follows   []string `json:"follows"`
	Followers []string `json:"followers"`
}

const PPActivitiesDir = "pp_activities_db"

type PPActivity struct {
	Posts []string `json:"posts"`
	Feed  []string `json:"feed"`
}

const PPPostsDir = "pp_posts_db"

type PPPost struct {
	Pseudo        string    `json:"pseudo"`
	Date          time.Time `json:"date"`
	Content       string    `json:"content"`
	DistributedTo []string  `json:"distributed_to"`
}

func NewPPEngine(root string) (*PPEngine, error) {
	usersDB, err := pebble.Open(path.Join(root, PPUsersDir), &pebble.Options{})
	if err != nil {
		return nil, err
	}

	followDB, err := pebble.Open(path.Join(root, PPFollowsDir), &pebble.Options{})
	if err != nil {
		return nil, err
	}

	activitesDB, err := pebble.Open(path.Join(root, PPActivitiesDir), &pebble.Options{})
	if err != nil {
		return nil, err
	}

	postsDB, err := pebble.Open(path.Join(root, PPPostsDir), &pebble.Options{})
	if err != nil {
		return nil, err
	}

	return &PPEngine{
		rootDir:      root,
		usersDB:      usersDB,
		followsDB:    followDB,
		activitiesDB: activitesDB,
		postsDB:      postsDB,
	}, nil
}

func (p *PPEngine) LogUser(pseudo string) error {
	key := []byte(pseudo)

	val, closer, err := p.usersDB.Get(key)
	if err == pebble.ErrNotFound {
		if err = p.createUser(pseudo); err != nil {
			return fmt.Errorf("failed to create user: pseudo=%s: %s", pseudo, err.Error())
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("failed request UsersDB: pseudo=%s: %s", pseudo, err.Error())
	}

	// Unmarshall current value
	var user *PPUser
	if err = json.Unmarshal(val, user); err != nil {
		return err
	}
	closer.Close()

	// Update LastConnect
	user.LastConnect = time.Now()
	if val, err = json.Marshal(user); err != nil {
		return err
	}

	// Update DB Entry
	if err = p.usersDB.Set(key, val, pebble.Sync); err != nil {
		return fmt.Errorf("failed to set user in UsersDB: pseudo=%s: %s\n", pseudo, err.Error())
	}

	// Insert value of the last loaded post
	p.minPostFeed[pseudo] = -1

	return nil
}

func (p *PPEngine) LogoutUser(pseudo string) error {
	delete(p.minPostFeed, pseudo)
	return nil
}

func (p *PPEngine) createUser(pseudo string) error {
	var val []byte
	var err error
	key := []byte(pseudo)

	// Users Entry
	user := &PPUser{
		Pseudo:      pseudo,
		LastConnect: time.Now(),
	}
	if val, err = json.Marshal(user); err != nil {
		return err
	}
	if err = p.usersDB.Set(key, val, pebble.Sync); err != nil {
		return err
	}

	// Follow Entry
	follow := &PPFollow{
		Follows:   []string{},
		Followers: []string{},
	}
	if val, err = json.Marshal(follow); err != nil {
		return err
	}
	if err = p.usersDB.Set(key, val, pebble.Sync); err != nil {
		return err
	}

	// Activity Entry
	activity := &PPActivity{
		Posts: []string{},
		Feed:  []string{},
	}
	if val, err = json.Marshal(activity); err != nil {
		return err
	}
	if err = p.usersDB.Set(key, val, pebble.Sync); err != nil {
		return err
	}

	return nil
}

func (p *PPEngine) Follow(src string, dst string) error {
	srcKey := []byte(src)
	dstKey := []byte(dst)

	// Direct Link
	val, closer, err := p.followsDB.Get(srcKey)
	if err != nil {
		return fmt.Errorf("failed to find follow entry: pseudo=%s: %s", src, err.Error())
	}

	var srcFollow *PPFollow
	if err = json.Unmarshal(val, srcFollow); err != nil {
		return err
	}
	closer.Close()

	srcFollow.Follows = append(srcFollow.Follows, dst)
	if val, err = json.Marshal(srcFollow); err != nil {
		return err
	}

	if err = p.followsDB.Set(srcKey, val, pebble.Sync); err != nil {
		return fmt.Errorf("failed to set follow entry: pseudo=%s: %s", src, err.Error())
	}

	// Reverse Link
	val, closer, err = p.followsDB.Get(dstKey)
	if err != nil {
		return fmt.Errorf("failed to find follow entry: pseudo=%s: %s", dst, err.Error())
	}

	var dstFollow *PPFollow
	if err = json.Unmarshal(val, dstFollow); err != nil {
		return err
	}
	closer.Close()

	dstFollow.Follows = append(dstFollow.Follows, src)
	if val, err = json.Marshal(dstFollow); err != nil {
		return err
	}

	if err = p.followsDB.Set(dstKey, val, pebble.Sync); err != nil {
		return fmt.Errorf("failed to set follow entry: pseudo=%s: %s", dst, err.Error())
	}

	return nil
}

func (p *PPEngine) UnFollow(src string, dst string) error {
	srcKey := []byte(src)
	dstKey := []byte(dst)

	// Direct link
	val, closer, err := p.followsDB.Get(srcKey)
	if err != nil {
		return fmt.Errorf("failed to find follow entry: pseudo=%s: %s", src, err.Error())
	}

	var srcFollow *PPFollow
	if err = json.Unmarshal(val, srcFollow); err != nil {
		return err
	}
	closer.Close()

	var idToDelete int
	for id, val := range srcFollow.Follows {
		if val == dst {
			idToDelete = id
			break
		}
	}
	srcFollow.Follows = append(srcFollow.Follows[:idToDelete], srcFollow.Follows[idToDelete+1:]...)

	if val, err = json.Marshal(srcFollow); err != nil {
		return err
	}

	if err = p.followsDB.Set(srcKey, val, pebble.Sync); err != nil {
		return fmt.Errorf("failed to set follow entry: pseudo=%s: %s", src, err.Error())
	}

	// Reverse link
	val, closer, err = p.followsDB.Get(dstKey)
	if err != nil {
		return fmt.Errorf("failed to find follow entry: pseudo=%s: %s", dst, err.Error())
	}

	var dstFollow *PPFollow
	if err = json.Unmarshal(val, dstFollow); err != nil {
		return err
	}
	closer.Close()

	for id, val := range dstFollow.Follows {
		if val == src {
			idToDelete = id
			break
		}
	}
	dstFollow.Follows = append(dstFollow.Follows[:idToDelete], dstFollow.Follows[idToDelete+1:]...)

	if val, err = json.Marshal(dstFollow); err != nil {
		return err
	}

	if err = p.followsDB.Set(dstKey, val, pebble.Sync); err != nil {
		return fmt.Errorf("failed to set follow entry: pseudo=%s: %s", dst, err.Error())
	}

	return nil
}

func (p *PPEngine) Post(pseudo string, content string) error {
	now := time.Now()

	userKey := []byte(pseudo)

	postId := fmt.Sprintf("%s:%d", pseudo, now.Unix())
	postKey := []byte(postId)

	// Find all the followers
	val, closer, err := p.followsDB.Get(userKey)
	if err != nil {
		return fmt.Errorf("failed to find follow entry: pseudo=%s: %s", pseudo, err.Error())
	}

	var follow *PPFollow
	if err = json.Unmarshal(val, follow); err != nil {
		return err
	}
	closer.Close()

	// Create the post
	post := &PPPost{
		Pseudo:        pseudo,
		Date:          now,
		Content:       content,
		DistributedTo: follow.Followers,
	}

	if val, err = json.Marshal(post); err != nil {
		return err
	}

	if err = p.postsDB.Set(postKey, val, pebble.Sync); err != nil {
		return fmt.Errorf("failed to set post entry: pseudo=%s: %s", pseudo, err.Error())
	}

	// Update all the feeds
	for _, psd := range follow.Followers {
		if err = p.appendFeed(psd, postId); err != nil {
			return err
		}
	}

	// Update user's feed
	val, closer, err = p.activitiesDB.Get(userKey)
	if err != nil {
		return fmt.Errorf("failed to get activity entry: pseudo=%s: %s", pseudo, err.Error())
	}

	var activity *PPActivity
	if err = json.Unmarshal(val, activity); err != nil {
		return err
	}
	closer.Close()

	activity.Posts = append(activity.Posts, postId)
	activity.Feed = append(activity.Feed, postId)

	if val, err = json.Marshal(activity); err != nil {
		return err
	}

	if err = p.activitiesDB.Set(userKey, val, pebble.Sync); err != nil {
		return fmt.Errorf("failed to set activity: pseudo=%s: %s", pseudo, err.Error())
	}

	return nil
}

func (p *PPEngine) appendFeed(pseudo string, postId string) error {
	userKey := []byte(pseudo)

	val, closer, err := p.activitiesDB.Get(userKey)
	if err != nil {
		return fmt.Errorf("failed to get activity entry: pseudo=%s: %s", pseudo, err.Error())
	}

	var activity *PPActivity
	if err = json.Unmarshal(val, activity); err != nil {
		return err
	}
	closer.Close()

	activity.Feed = append(activity.Feed, postId)
	if val, err = json.Marshal(activity); err != nil {
		return err
	}

	if err = p.activitiesDB.Set(userKey, val, pebble.Sync); err != nil {
		return fmt.Errorf("failed to set activity: pseudo=%s: %s", pseudo, err.Error())
	}

	return nil
}

func (p *PPEngine) GetFeed(pseudo string) ([]*PPPost, error) {
	userKey := []byte(pseudo)

	val, closer, err := p.activitiesDB.Get(userKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity entry: pseudo=%s: %s", pseudo, err.Error())
	}

	var activity *PPActivity
	if err = json.Unmarshal(val, activity); err != nil {
		return nil, err
	}
	closer.Close()

	startId := len(activity.Feed) - 1
	if p.minPostFeed[pseudo] > 0 {
		startId = p.minPostFeed[pseudo]
	}

	var posts []*PPPost
	for id := startId; id >= 0; id -= 1 {
		var post *PPPost

		pid := activity.Feed[id]
		if post, err = p.getPost(pid); err != nil {
			return nil, err
		}
		posts = append(posts, post)

		if len(posts) == 10 {
			p.minPostFeed[pseudo] = id
			break
		}
	}

	return posts, nil
}

func (p *PPEngine) getPost(postId string) (*PPPost, error) {
	postKey := []byte(postId)

	val, closer, err := p.postsDB.Get(postKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get post entry: postId=%s: %s", postId, err.Error())
	}

	var post *PPPost
	if err = json.Unmarshal(val, post); err != nil {
		return nil, err
	}
	closer.Close()

	return post, nil
}
