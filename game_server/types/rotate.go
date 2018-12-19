package types

// User1 видит карту так же, как и сервер. User0 видит всё отражённым,
// что бы небыло различий между пользователями на фронте.
// приходяшие структуры надо разворачивать сразу после парсинга, 1 раз.

func (a *AttemptGoToCell) Rotate() {
	a.From = 41 - a.From
	a.To = 41 - a.To
	return
}

func (rw *ReassignWeapons) Rotate() {
	rw.CharacterPosition = 41 - rw.CharacterPosition
	return
}

func (dm *DownloadMap) Rotate() {
	for i := 0; i < 21; i++ {
		dm[i], dm[41-i] = dm[41-i], dm[i]
	}
	return
}

func (mc *MoveCharacter) Rotate() {
	mc.From = 41 - mc.From
	mc.To = 41 - mc.To
	return
}

func (a *Attack) Rotate() {
	a.Winner.Coordinates = 41 - a.Winner.Coordinates
	a.Loser.Coordinates = 41 - a.Loser.Coordinates
	return
}

func (aw *AddWeapon) Rotate() {
	aw.Coordinates = 41 - aw.Coordinates
	return
}

func (wcr *WeaponChangeRequest) Rotate() {
	wcr.CharacterPosition = 41 - wcr.CharacterPosition
	return
}

func (g *GameOver) Rotate() {
	g.From = 41 - g.From
	g.To = 41 - g.To
	return
}
