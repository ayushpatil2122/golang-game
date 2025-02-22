package main

import (
	"github.com/hajimehoshi/ebiten"
	"fmt"
	"image/png"
	"os"
	"strconv"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
	"github.com/hajimehoshi/ebiten/inpututil"
)


type Player struct {
	//attributes
	Level          int
	EXP            int
	Max_EXP        int
	ATK            int
	MAX_HP, MAX_MP int
	HP, MP         int //current HP and MP
	Money          int
	Step           float64 //about walking, jumping and spell speed(pix per frame)
	//to draw player pic
	Point                              //location in screen
	Left      bool                     //false: player faces right; true: player faces left
	Pics      map[string]*ebiten.Image //the index of the map is the name of the image
	Pic_index string                   //the index of the player pic map
	//status
	Jumping   bool
	Attacking bool
	//weapon and spell
	weapon Weapon
	spells []Spell //recently released spells
}

type Weapon struct {
	Point
	Pics      map[string]*ebiten.Image
	Pic_index string
}

func (player *Player) Init() {
	//initial attributes
	player.Level = 1
	player.Max_EXP = 100
	player.ATK = 1
	player.Step = 4
	player.MAX_HP = 100
	player.HP = 100
	player.MAX_MP = 100
	player.MP = 100
	player.Money = 10
	//set initial location, located at left(1/3), upon ground
	player.X = (screen_width - player_width) / 3
	player.Y = screen_height - player_height - ground_height
	//make map
	player.Pics = make(map[string]*ebiten.Image)
	//initial pic is stand
	player.Pic_index = "player_stand.png"
	//init weapon
	player.weapon.X = player.X + player_width
	player.weapon.Y = player.Y
	player.weapon.Pics = make(map[string]*ebiten.Image)
	player.weapon.Pic_index = "weapon_standby.png"
	// HP and MP recovery
	go player.Recovery()
}

func (player *Player) Get_Movement() {
	/*get player input and do the movement*/
	//walk
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyD) {
		if ebiten.IsKeyPressed(ebiten.KeyA) {
			player.Walk(ebiten.KeyA)
		}
		if ebiten.IsKeyPressed(ebiten.KeyD) {
			player.Walk(ebiten.KeyD)
		}
	}
	//jump
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyK) {
		go player.Jump() //use another goroutine to jump
	}
	//attack, update weapon info
	player.Attack()
	//cast spell thing
	if inpututil.IsKeyJustPressed(ebiten.KeyU) {
		player.Cast_Spell_U()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		player.Cast_Spell_I()
	}
	player.Update_spell()
	//if level up
	if player.EXP >= player.Max_EXP {
		player.Upgrade()
	}
}

func (player *Player) Walk(key ebiten.Key) {
	const border_distance = 3
	//update loaction
	if key == ebiten.KeyA {
		player.Left = true
		process.distance -= player.Step
		player.X -= player.Step
		if player.X < player_width*border_distance {
			//player should be at least 3 person'width away from the left border
			player.X = player_width * border_distance
			//when player gets the border, plus others'X, instead minus player'X,
			//like player is still moving
			background.Player_Moving()
			monsters.Player_Moving()
			coins.Player_Moving()
			shop.Player_Moving()
		}
	} else if key == ebiten.KeyD {
		player.Left = false
		process.distance += player.Step
		player.X += player.Step
		if player.X > screen_width-player_width*(border_distance+1) {
			//right border too(*4 cuz drawing point is left-up point )
			player.X = screen_width - player_width*(border_distance+1)
			//right border, same way to deal it
			background.Player_Moving()
			monsters.Player_Moving()
			coins.Player_Moving()
			shop.Player_Moving()
		}
	}
	//change leg pic
	if process.frame_cnt%3 == 0 { //every 3 frame, change leg
		if player.Pic_index == "player_walk1.png" || player.Pic_index == "player_walk1_L.png" {
			if player.Left {
				player.Pic_index = "player_walk2_L.png"
			} else {
				player.Pic_index = "player_walk2.png"
			}
		} else if player.Pic_index == "player_walk2.png" || player.Pic_index == "player_walk2_L.png" {
			if player.Left {
				player.Pic_index = "player_walk1_L.png"
			} else {
				player.Pic_index = "player_walk1.png"
			}
		} else { //player was not walking just now
			//pick a random walk pic
			if rand.Intn(2) == 1 {
				if player.Left {
					player.Pic_index = "player_walk1_L.png"
				} else {
					player.Pic_index = "player_walk1.png"
				}
			} else {
				if player.Left {
					player.Pic_index = "player_walk2_L.png"
				} else {
					player.Pic_index = "player_walk2.png"
				}
			}
		}
	}
	//if player eats coins
	coins = coins.Player_get()
}

func (player *Player) Jump() {
	const player_jump_height = 200 //player highest jump height

	if !player.Jumping {
		player.Jumping = true //lock, refuse more jump signal
		//jumping pic
		if player.Left {
			player.Pic_index = "player_jumping_L.png"
		} else {
			player.Pic_index = "player_jumping.png"
		}
		//jumping
		t := (player_jump_height / player.Step) / 2
		for i := 0; i < int(t); i++ { //up
			player.Y -= player.Step * 2
			time.Sleep(time.Second / 60)
		}
		for i := 0; i < int(t); i++ { //down
			player.Y += player.Step * 2
			time.Sleep(time.Second / 60)
		}

		player.Y = screen_height - player_height - ground_height //make sure player back to ground
		player.Jumping = false                                   //unlock
	}
}

func (player *Player) Attack() {
	//pick player and weapon pic
	if ebiten.IsKeyPressed(ebiten.KeyJ) { //player is attacking
		player.Attacking = true
		if player.Left {
			player.weapon.Pic_index = "weapon_attack_L.png"
			player.Pic_index = "player_stand_L.png"
		} else {
			player.weapon.Pic_index = "weapon_attack.png"
			player.Pic_index = "player_stand.png"
		}
		//if hit monster
		for i := 0; i < len(monsters); i++ {
			if overlap(player.weapon.Pics[player.weapon.Pic_index], monsters[i].Pic[monsters[i].Pic_index], player.weapon.Point, monsters[i].Point) {
				monsters[i].HP -= player.ATK
				//knockback
				if player.Left {
					monsters[i].X -= (monsters[i].Step + 1)
				} else {
					monsters[i].X += (monsters[i].Step + 1)
				}
			}
		}
	} else { //player is not attacking
		player.Attacking = false
		//draw weapon standby pic
		if player.Left {
			player.weapon.Pic_index = "weapon_standby_L.png"
		} else {
			player.weapon.Pic_index = "weapon_standby.png"
		}
	}
	//update weapon location
	if player.Left {
		player.weapon.X = player.X - weapon_size
		player.weapon.Y = player.Y
	} else {
		player.weapon.X = player.X + player_width
		player.weapon.Y = player.Y
	}
}

func (player *Player) Cast_Spell_U() {
	const ATK_rate_U = 3
	const MP_cost = 10
	/* A horizontal forward ball */
	left := player.Left //Mark player's current direction
	//player casting spell pic
	if left {
		player.Pic_index = "player_spell_L.png"
	} else {
		player.Pic_index = "player_spell.png"
	}
	//add new spell to slice
	if player.MP >= MP_cost {
		player.MP -= MP_cost

		var new_spell Spell
		new_spell.Name = "spell_U"
		new_spell.ATK_rate = ATK_rate_U
		new_spell.Pic = get_img(new_spell.Name + ".png")
		new_spell.Y = player.Y + 54 //by my experiment, this height looks well
		if left {
			new_spell.X = player.X - spell_U_size
			new_spell.direction = "left"
		} else {
			new_spell.X = player.X + player_width
			new_spell.direction = "right"
		}
		player.spells = append(player.spells, new_spell)
	}
}

func (player *Player) Cast_Spell_I() {
	/*a stuff goes straight down from the sky*/
	const ATK_rate_I = 10
	const MP_cost = 20
	left := player.Left
	if left {
		player.Pic_index = "player_spell_L.png"
	} else {
		player.Pic_index = "player_spell.png"
	}
	if player.MP >= MP_cost {
		player.MP -= MP_cost

		var new_spell Spell
		new_spell.ATK_rate = ATK_rate_I
		new_spell.Name = "spell_I"
		new_spell.Pic = get_img(new_spell.Name + ".png")
		new_spell.Y = -spell_I_height
		new_spell.direction = "down"
		//spell is two body width away from player
		if left {
			new_spell.X = player.X - 3*player_width
		} else {
			new_spell.X = player.X + 2*player_width
		}
		player.spells = append(player.spells, new_spell)
	}
}

func (player *Player) Update_spell() {
	var spell_step float64 = player.Step + 3

	for i := 0; i < len(player.spells); i++ {
		//update location
		if player.spells[i].direction == "left" {
			player.spells[i].X -= spell_step
		} else if player.spells[i].direction == "right" {
			player.spells[i].X += spell_step
		} else if player.spells[i].direction == "down" {
			player.spells[i].Y += spell_step
		}
		//delete spell out of bounds
		if player.spells[i].X < (-spell_U_size) || player.spells[i].X > (screen_width) {
			player.spells = append(player.spells[:i], player.spells[i+1:]...)
			//to avoid index out of range
			i--
			break
		}
		if player.spells[i].Y > screen_height {
			player.spells = append(player.spells[:i], player.spells[i+1:]...)
			break
		}
		//if hit monster
		for j := 0; j < len(monsters); j++ {
			if overlap(player.spells[i].Pic, monsters[j].Pic[monsters[j].Pic_index], player.spells[i].Point, monsters[j].Point) {
				monsters[j].HP -= player.spells[i].ATK_rate * player.ATK
			}
		}
	}
}

func (player *Player) Upgrade() {
	player.Level++
	player.MAX_HP += 100
	player.MAX_MP += 100
	player.ATK = player.Level
	player.EXP = 0
}

func (player *Player) Recovery() {
	for player.HP > 0 {
		if player.HP += player.MAX_HP / 100; player.HP > player.MAX_HP {
			player.HP = player.MAX_HP
		}
		if player.MP += player.MAX_MP / 100; player.MP > player.MAX_MP {
			player.MP = player.MAX_MP
		}
		time.Sleep(time.Second)
	}
}

type Monster struct {
	Name string
	//attributes
	Level      int
	ATK        int
	MAX_HP, HP int
	Step       float64 //pix per frame
	// to draw pic
	Point
	Self_Point Point //Take birth point as origin of coordinate system
	Left       bool
	Pic        map[string]*ebiten.Image
	Pic_index  string
	//status
	Attacking bool
	//attack thing
	Spell
}

type Monsters []Monster

func (monsters Monsters) Update() Monsters {
	monsters = monsters.Birth()
	monsters.Walk()
	monsters.If_Attack()
	monsters.Update_Spell()
	monsters = monsters.Death()
	return monsters
}

func (monsters Monsters) Birth() Monsters {
	const distance_interval = 1800
	//Every (distance_interval) pixes player walks, a group of monsters will be born
	const birth_zone = 1200 //monster will birth in the zone of [x,x+BirthZone), x is the screen_width
	const max_number = 7    //number of monsters at most in a group
	const pic_num = 2

	if int(process.distance/distance_interval) == process.monster_group_cnt {
		//create a group of mosnters
		num := rand.Intn(max_number) + 1
		for i := 0; i < num; i++ {
			//init a monster
			var monster Monster
			monster.Name = "monster_" + strconv.Itoa((process.stage+pic_num)%pic_num)
			monster.Step = 2
			monster.X = float64(rand.Intn(birth_zone) + screen_width)
			monster.Y = screen_height - ground_height - monster_size
			monster.Left = (rand.Intn(2) == 1) //randon direction
			monster.Pic = make(map[string]*ebiten.Image)
			monster.Pic[monster.Name+"_L.png"] = get_img(monster.Name + "_L.png")
			monster.Pic[monster.Name+".png"] = get_img(monster.Name + ".png")
			if monster.Left {
				monster.Pic_index = monster.Name + "_L.png"
			} else {
				monster.Pic_index = monster.Name + ".png"
			}
			monster.ATK = (process.stage)
			monster.MAX_HP = (process.stage) * 100
			monster.HP = monster.MAX_HP
			//add it to slice
			monsters = append(monsters, monster)
		}
		process.monster_group_cnt++
	}
	return monsters
}

func (monsters Monsters) Walk() {
	const Max_Distance = 300 //to make monster active zone
	for i := 0; i < len(monsters); i++ {
		if monsters[i].Left {
			monsters[i].X -= monsters[i].Step
			monsters[i].Self_Point.X -= monsters[i].Step
			monsters[i].Pic_index = monsters[i].Name + "_L.png"
			if monsters[i].Self_Point.X < (-Max_Distance) {
				//change direction
				monsters[i].Left = false
			}
		} else {
			monsters[i].X += monsters[i].Step
			monsters[i].Self_Point.X += monsters[i].Step
			monsters[i].Pic_index = monsters[i].Name + ".png"
			if monsters[i].Self_Point.X > Max_Distance {
				monsters[i].Left = true
			}
		}
	}
}

func (monsters Monsters) If_Attack() {
	const attack_chance = 500 // 1/500 chance to attack
	for i, _ := range monsters {
		if rand.Intn(attack_chance) == 0 && !monsters[i].Attacking {
			monsters[i].Attacking = true
			monsters[i].Spell.Pic = get_img("monster_spell.png")
			monsters[i].Spell.mark = process.frame_cnt
		}
	}
}

func (monsters Monsters) Update_Spell() {
	const spell_duration = 60 //(frames)
	const spell_step = 1
	for i, _ := range monsters {
		if monsters[i].Attacking {
			if process.frame_cnt < monsters[i].Spell.mark+spell_duration {
				if monsters[i].Left {
					monsters[i].Spell.X = monsters[i].X - mosnter_spell_size - spell_step*(float64(process.frame_cnt-monsters[i].Spell.mark))
					monsters[i].Spell.Y = monsters[i].Y + (monster_size-mosnter_spell_size)/2
				} else {
					monsters[i].Spell.X = monsters[i].X + monster_size + spell_step*(float64(process.frame_cnt-monsters[i].Spell.mark))
					monsters[i].Spell.Y = monsters[i].Y + (monster_size-mosnter_spell_size)/2
				}
				if overlap(monsters[i].Spell.Pic, player.Pics[player.Pic_index], monsters[i].Spell.Point, player.Point) {
					player.HP -= monsters[i].ATK
				}
			} else { //spell time is over
				monsters[i].Attacking = false

			}
		}
	}
}

func (monsters Monsters) Player_Moving() {
	for i, _ := range monsters {
		if player.Left {
			monsters[i].X += player.Step
		} else {
			monsters[i].X -= player.Step
		}
	}
}

func (monsters Monsters) Death() Monsters {
	//walk monsters, if any dies, transform into coin
	for i := 0; i < len(monsters); i++ {
		if monsters[i].HP <= 0 {
			//turn into coin
			var coin Coin
			coin.X = monsters[i].X + (monster_size-coin_size)/2
			coin.Y = screen_height - ground_height - coin_size
			coin.img = get_img("coin.png")
			coin.value = process.stage
			coins = append(coins, coin)
			//delete died one
			monsters = append(monsters[:i], monsters[i+1:]...)
			//player gain exp
			player.EXP += 15
		}
	}
	return monsters
}

type Coin struct {
	Point
	value int
	img   *ebiten.Image
}

type Coins []Coin

type Shop struct {
	Created int
	Point
	img        *ebiten.Image
	cost       int
	HP_growth  int
	ATK_growth int
}

func (coins Coins) Player_Moving() {
	for i, _ := range coins {
		if player.Left {
			coins[i].X += player.Step
		} else {
			coins[i].X -= player.Step
		}
	}
}

func (coins Coins) Player_get() Coins {
	for i := 0; i < len(coins); i++ {
		if overlap(player.Pics[player.Pic_index], coins[i].img, player.Point, coins[i].Point) {
			player.Money += coins[i].value
			coins = append(coins[:i], coins[i+1:]...)
		}
	}
	return coins
}

func (shop *Shop) Update() {
	shop.Create()
	shop.Buying()
}

func (shop *Shop) Create() {
	const create_duration = 2000
	if int(process.distance/create_duration) == shop.Created {
		shop.Created++
		//show shop in screen
		shop.Y = screen_height - player_height - ground_height - 100
		shop.X = screen_width
		shop.img = get_img("shop.png")
		//get random cost and growth
		shop.cost = rand.Intn(process.stage * 17)
		shop.HP_growth = rand.Intn(process.stage * 50)
		shop.ATK_growth = rand.Intn(process.stage * 2)

		ebitenutil.DebugPrint(shop.img, "   Money-"+strconv.Itoa(shop.cost))
		ebitenutil.DebugPrint(shop.img, "\n   ATK+"+strconv.Itoa(shop.ATK_growth))
		ebitenutil.DebugPrint(shop.img, "\n\n   MAX_HP+"+strconv.Itoa(shop.HP_growth))
	}
}

func (shop *Shop) Buying() {
	if overlap(player.Pics[player.Pic_index], shop.img, player.Point, shop.Point) {
		if player.Money >= shop.cost {
			player.Money -= shop.cost
			player.MAX_HP += shop.HP_growth
			player.ATK += shop.ATK_growth
			//delete shop then
			shop.Y = -100
		}
	}
}

func (shop *Shop) Player_Moving() {
	if player.Left {
		shop.X += player.Step
	} else {
		shop.X -= player.Step
	}
}

func Init() {
	rand.Seed(time.Now().Unix())

	player.Init()
	background.Init()

	load_pics()

}

func load_pics() {
	files, err := ioutil.ReadDir("resources/pics") //get a slice of files info
	handle_error(err)
	for _, pic := range files {
		img := get_img(pic.Name()) //get the img
		//put the img to right place
		if strings.Index(pic.Name(), "player") != -1 { //it's a player pic
			player.Pics[pic.Name()] = img //add it to player pics map
		} else if strings.Index(pic.Name(), "back") != -1 {
			background.back_pic = img
		} else if strings.Index(pic.Name(), "ground") != -1 {
			background.ground_pic = img
		} else if strings.Index(pic.Name(), "weapon") != -1 {
			player.weapon.Pics[pic.Name()] = img
		} else if strings.Index(pic.Name(), "hit") != -1 {
			hit_pics.Pic = img
		}
	}
}

type Process struct {
	//game process related
	distance  float64 //how far has player gone. it influences stage
	stage     int     //it's about game difficulty, monsters level and something like that
	frame_cnt int     //timer, to help to create some duration. it plus one every frame
	//background pics
	monster_group_cnt int
	pause             bool
}

func (process *Process) Update() {
	const stage_duration = 2000
	process.frame_cnt++
	if process.frame_cnt > 2000000000 {
		process.frame_cnt = 0 //avoid overflow
	}
	process.stage = int(process.distance/stage_duration) + 1
}

func Draw(screen *ebiten.Image) {
	var op ebiten.DrawImageOptions
	//draw background
	op.GeoM.Translate(0, 0)
	screen.DrawImage(background.back_pic, &op)
	op.GeoM.Reset()
	for _, ground := range background.ground_point {
		op.GeoM.Translate(ground.X, ground.Y)
		screen.DrawImage(background.ground_pic, &op)
		op.GeoM.Reset()
	}
	//draw monsters
	for _, monster := range monsters {
		if monster.X > (-monster_size) && monster.X < screen_width { //draw monsters in the screen scope only
			op.GeoM.Translate(monster.X, monster.Y)
			screen.DrawImage(monster.Pic[monster.Pic_index], &op)
			op.GeoM.Reset()
		}
	}
	//draw monster spells(if attacking)
	for _, monster := range monsters {
		if monster.X > (-monster_size) && monster.X < screen_width {
			if monster.Attacking {
				op.GeoM.Translate(monster.Spell.X, monster.Spell.Y)
				screen.DrawImage(monster.Spell.Pic, &op)
				op.GeoM.Reset()
			}
		}
	}
	//draw coins
	for _, coin := range coins {
		op.GeoM.Translate(coin.X, coin.Y)
		screen.DrawImage(coin.img, &op)
		op.GeoM.Reset()
	}
	//draw shop
	op.GeoM.Translate(shop.X, shop.Y)
	screen.DrawImage(shop.img, &op)
	op.GeoM.Reset()
	//draw player
	op.GeoM.Translate(player.X, player.Y)
	screen.DrawImage(player.Pics[player.Pic_index], &op) //draw player pic according index
	op.GeoM.Reset()
	//draw weapon
	if player.Pic_index != "player_spell.png" && player.Pic_index != "player_spell_L.png" {
		//player is not casting spell,
		op.GeoM.Translate(player.weapon.X, player.weapon.Y)
		screen.DrawImage(player.weapon.Pics[player.weapon.Pic_index], &op)
		op.GeoM.Reset()
	}
	//draw spells
	for _, spell := range player.spells {
		op.GeoM.Translate(spell.X, spell.Y)
		screen.DrawImage(spell.Pic, &op)
		op.GeoM.Reset()
	}
	//draw hit pics
	for i := 0; i < len(hit_pics.Points); i++ {
		op.GeoM.Translate(hit_pics.Points[i].X, hit_pics.Points[i].Y)
		screen.DrawImage(hit_pics.Pic, &op)
		op.GeoM.Reset()
	}
	//delete hit pics points
	hit_pics.Points = hit_pics.Points[0:0]
	//print player state
	ebitenutil.DebugPrint(screen, "lv: "+strconv.Itoa(player.Level)+"   money: "+strconv.Itoa(player.Money)+"   ATK: "+strconv.Itoa(player.ATK))
	ebitenutil.DebugPrint(screen, "\nHP: "+strconv.Itoa(player.HP)+"/"+strconv.Itoa(player.MAX_HP)+"   MP "+strconv.Itoa(player.MP)+"/"+strconv.Itoa(player.MAX_MP))
}

const (
	screen_width       = 1200
	screen_height      = 600
	ground_width       = 1200
	ground_height      = 100
	player_width       = 108
	player_height      = 150
	weapon_size        = 150 
	spell_U_size       = 36
	spell_I_width      = 200
	spell_I_height     = 100
	monster_size       = 120
	mosnter_spell_size = 60
	hit_pic_size       = 20
	coin_size          = 50
)

var (
	process    Process
	background Background
	player     Player
	monsters   Monsters
	hit_pics   Hit_Pics
	coins      Coins
	shop       Shop
)

type Point struct {
	X, Y float64
}

type Spell struct {
	Name     string
	ATK_rate int // =  player_akt * atk_rate
	MP_cost  int
	Point
	direction string
	Pic       *ebiten.Image
	mark      int //to mark frame_count
}

type Hit_Pics struct {
	Points []Point
	Pic    *ebiten.Image
}

func handle_error(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func get_img(name string) *ebiten.Image {
	file, err := os.Open("resources/pics/" + name) //read file
	handle_error(err)
	defer file.Close()
	//decode file to image.Image
	src_img, err := png.Decode(file)
	handle_error(err)
	// turn image.Image into *ebiten.Image
	img, _ := ebiten.NewImageFromImage(src_img, ebiten.FilterDefault)
	//this function always return nil error, so I Ignore it
	return img
}

func overlap(a, b *ebiten.Image, pa, pb Point) bool {
	if (pa.X < pb.X+float64(b.Bounds().Dx())) && (pa.X+float64(a.Bounds().Dx()) > pb.X) {
		if (pa.Y < pb.Y+float64(b.Bounds().Dy())) && (pa.Y+float64(a.Bounds().Dy()) > pb.Y) {
			add_hit_pic(a, b, pa, pb)
			return true
		}
	}
	return false
}

func add_hit_pic(a, b *ebiten.Image, pa, pb Point) {
	var p Point
	center_a_x := (pa.X + pa.X + float64(a.Bounds().Dx())) / 2
	center_a_y := (pa.Y + pa.Y + float64(a.Bounds().Dy())) / 2
	center_b_x := (pb.X + pb.X + float64(b.Bounds().Dx())) / 2
	center_b_y := (pb.Y + pb.Y + float64(b.Bounds().Dy())) / 2
	p.X = (center_a_x+center_b_x)/2 - hit_pic_size/2
	p.Y = (center_a_y+center_b_y)/2 - hit_pic_size/2
	hit_pics.Points = append(hit_pics.Points, p)
}

type Background struct {
	//background pics
	back_pic     *ebiten.Image
	ground_pic   *ebiten.Image
	ground_point []Point //there should almost always two ground pics in the screen
}

func (background *Background) Init() {
	var ground Point
	ground.X = 0
	ground.Y = screen_height - ground_height
	background.ground_point = append(background.ground_point, ground)
}

func (background *Background) Player_Moving() {
	if player.Left {
		//create a new ground to adjoin the old one in the left, like they are whole
		if len(background.ground_point) == 1 {
			//there are 2 ground pic in screen, if only one ground there, create a new one
			var ground Point
			ground.X = background.ground_point[0].X - ground_width
			ground.Y = background.ground_point[0].Y
			background.ground_point = append(background.ground_point, ground)
		}
		//player is walking to left, increase ground'X
		for i, _ := range background.ground_point {
			background.ground_point[i].X += player.Step
		}
	} else {
		if len(background.ground_point) == 1 {
			var ground Point
			ground.X = background.ground_point[0].X + ground_width
			ground.Y = background.ground_point[0].Y
			background.ground_point = append(background.ground_point, ground)
		}
		for i, _ := range background.ground_point {
			background.ground_point[i].X -= player.Step
		}
	}
	//delete ground pic out of bounds
	for i := 0; i < len(background.ground_point); i++ {
		if background.ground_point[i].X < (-ground_width) || background.ground_point[i].X > ground_width {
			background.ground_point = append(background.ground_point[:i], background.ground_point[i+1:]...)
		}
	}
}

func main() {
	Init()
	err := ebiten.Run(Game, screen_width, screen_height, 1, "Game")
	handle_error(err)
}

func Game(screen *ebiten.Image) error {
	if player.HP > 0 {
		//update process thing
		process.Update()
		//update player
		player.Get_Movement()
		//update monsters
		monsters = monsters.Update()
		//shop thing
		shop.Update()
	}
	//draw stuff
	Draw(screen)
	return nil
}