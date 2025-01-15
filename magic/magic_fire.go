package magic

import "fmt"

type fireMagic struct {
}

func NewFireMagic() Magic {
	return &fireMagic{}
}

// Name implements Magic.
func (f *fireMagic) Name() string {
	return "FireMagic"
}

// Trigger implements Magic.
func (f *fireMagic) Trigger() error {
	fmt.Printf("Magic %s is triggered\n\n", f.Name())
	return f.Expolosion()
}

func (f *fireMagic) Expolosion() error {
	fmt.Printf(`黒より黒く 闇より暗き漆黒に
我が真紅の混交を望み給う
覚醒の時来たれり 無謬の境界に落ちし理
むぎょうの歪みとなりて現出せよ！！
踊れ、踊れ、踊れ！
我が力の奔流に望むは
崩壊なり
並ぶものなき崩壊なり！
万象等しく灰燼にきし 深淵より来たれ！
これが、人類最大の威力の攻撃手段！
これこそが！究極の攻撃魔法！
`)
	return nil
}
