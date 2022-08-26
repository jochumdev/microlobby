package gamedb

import (
	"context"
	"net/http"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"go-micro.dev/v4/errors"
	"wz2100.net/microlobby/service/gamedb/v1/db"
	"wz2100.net/microlobby/service/gamedb/v1/version"
	"wz2100.net/microlobby/shared/auth"
	"wz2100.net/microlobby/shared/component"
	"wz2100.net/microlobby/shared/proto/gamedbpb/v1"
	"wz2100.net/microlobby/shared/utils"
)

const pkgPath = version.PkgPath + "/service/gamedb"

func dbPlayerToProto(dp *db.GamePlayer) (*gamedbpb.Player, error) {
	return &gamedbpb.Player{
		Uuid:      dp.UUID.String(),
		Name:      dp.Name,
		IpAddress: dp.IpAddress,
		IsHost:    dp.IsHost,
	}, nil
}

func dbGameToProto(dg *db.Game, pg *gamedbpb.Game) error {
	protoPlayers := []*gamedbpb.Player{}
	for _, dp := range dg.Players {
		protoPlayer, err := dbPlayerToProto(dp)
		if err != nil {
			return err
		}
		protoPlayers = append(protoPlayers, protoPlayer)
	}

	pg.Id = dg.Id.String()
	pg.Description = dg.Description
	pg.Map = dg.Map
	pg.Mods = dg.Mods
	pg.HostIp = dg.HostIp
	pg.Port = dg.Port
	pg.Players = protoPlayers
	pg.MaxPlayers = dg.MaxPlayers
	pg.Version = dg.Version
	pg.VerMajor = dg.VerMajor
	pg.VerMinor = dg.VerMinor
	pg.IsPure = dg.IsPure
	pg.IsPrivate = dg.IsPrivate

	pg.V3GameId = dg.V3GameId
	pg.LobbyVersion = dg.LobbyVersion

	return nil
}

func protoPlayerToDB(pp *gamedbpb.Player) (*db.GamePlayer, error) {
	if len(pp.Uuid) < 1 {
		return &db.GamePlayer{
			Name:      pp.Name,
			IpAddress: pp.IpAddress,
			IsHost:    pp.IsHost,
		}, nil
	}

	pUuid, err := uuid.Parse(pp.Uuid)
	if err != nil {
		return nil, err
	}
	return &db.GamePlayer{
		UUID:      pUuid,
		Name:      pp.Name,
		IpAddress: pp.IpAddress,
		IsHost:    pp.IsHost,
	}, nil
}

func protoGameToDB(pg *gamedbpb.Game, dg *db.Game) error {
	if len(pg.Id) > 0 {
		pUuid, err := uuid.Parse(pg.Id)
		if err != nil {
			return err
		}

		dg.Id = pUuid
	}

	players := []*db.GamePlayer{}
	for _, pp := range pg.Players {
		player, err := protoPlayerToDB(pp)
		if err != nil {
			return err
		}
		players = append(players, player)
	}

	dg.Description = pg.Description
	dg.Map = pg.Map
	dg.Mods = pg.Mods
	dg.HostIp = pg.HostIp
	dg.Port = pg.Port
	dg.Players = players
	dg.MaxPlayers = pg.MaxPlayers
	dg.Version = pg.Version
	dg.VerMajor = pg.VerMajor
	dg.VerMinor = pg.VerMinor
	dg.IsPure = pg.IsPure
	dg.IsPrivate = pg.IsPrivate

	dg.V3GameId = pg.V3GameId
	dg.LobbyVersion = pg.LobbyVersion

	return nil
}

type Handler struct {
	cRegistry *component.Registry
	logrus    component.LogrusComponent
	svcName   string
}

func NewHandler(cregistry *component.Registry) (*Handler, error) {
	h := &Handler{
		cRegistry: cregistry,
		svcName:   cregistry.Service.Name(),
	}

	return h, nil
}

func (h *Handler) Start() error {
	logrus, err := component.Logrus(h.cRegistry)
	if err != nil {
		return errors.FromError(err)
	}
	h.logrus = logrus

	return nil
}
func (h *Handler) Stop() error { return nil }

func (h *Handler) List(ctx context.Context, in *gamedbpb.ListRequest, out *gamedbpb.ListResponse) error {
	user, err := utils.CtxMetadataUser(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	if in.History {
		if !auth.IntersectsRoles(user, auth.AllowServiceAndAdmin...) {
			return errors.New(h.svcName, "Your not allowed to make history requests", http.StatusBadRequest)
		}
	}

	// Get the database engine
	bun, err := component.Bun(h.cRegistry)
	if err != nil {
		return errors.FromError(err)
	}

	count, err := bun.NewSelect().Model((*db.Game)(nil)).Count(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	var games []db.Game
	err = bun.NewSelect().
		Model(&games).
		Relation("Players").
		ColumnExpr("g.*").
		Limit(int(in.Limit)).
		Offset(int(in.Offset)).
		Scan(ctx)

	if err != nil {
		return errors.FromError(err)
	}

	out.Count = uint64(count)
	for _, g := range games {
		pg := &gamedbpb.Game{}
		err := dbGameToProto(&g, pg)
		if err != nil {
			return errors.FromError(err)
		}

		out.Games = append(out.Games, pg)
	}

	return nil
}

func (h *Handler) checkGame(ctx context.Context, dg *db.Game, oldG *db.Game) error {
	user, err := utils.CtxMetadataUser(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	if dg.V3GameId != 0 {
		if !auth.IntersectsRoles(user, auth.AllowServiceAndAdmin...) {
			return errors.New(h.svcName, "You'r not allowed to make V3GameID requests", http.StatusBadRequest)
		}

		if dg.LobbyVersion != 3 {
			return errors.New(h.svcName, "V3GameID is only allowed for LobbyVersion 3", http.StatusBadRequest)
		}
	}

	var nHostPlayer *db.GamePlayer
	for _, dp := range dg.Players {
		if dp.IsHost && dp.IpAddress == dg.HostIp {
			nHostPlayer = dp
			break
		}
	}
	if nHostPlayer == nil {
		return errors.New(h.svcName, "No host player given or IP's don't match", http.StatusBadRequest)
	}

	// Update?
	if oldG != nil {
		var oldHostPlayer *db.GamePlayer
		for _, dp := range oldG.Players {
			if dp.IsHost && dp.IpAddress == dg.HostIp {
				oldHostPlayer = dp
				break
			}
		}

		if oldHostPlayer == nil ||
			oldHostPlayer.UUID != nHostPlayer.UUID ||
			oldG.HostIp != dg.HostIp ||
			oldG.LobbyVersion != dg.LobbyVersion ||
			oldG.V3GameId != dg.V3GameId ||
			oldG.Version != dg.Version ||
			oldG.VerMajor != dg.VerMajor ||
			oldG.VerMinor != dg.VerMinor ||
			oldG.IsPure != dg.IsPure {
			return errors.New(h.svcName, "An update is not allowed", http.StatusMethodNotAllowed)
		}
	}

	return nil
}

func (h *Handler) Create(ctx context.Context, in *gamedbpb.Game, out *gamedbpb.Game) error {
	dg := &db.Game{}
	err := protoGameToDB(in, dg)
	if err != nil {
		return errors.FromError(err)
	}

	if err := h.checkGame(ctx, dg, nil); err != nil {
		return err
	}

	// Get the database engine
	bun, err := component.Bun(h.cRegistry)
	if err != nil {
		return errors.FromError(err)
	}

	_, err = bun.NewInsert().
		Model(dg).
		Exec(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	for _, dp := range dg.Players {
		dp.GameID = dg.Id
		_, err = bun.NewInsert().
			Model(dp).
			Exec(ctx)
		if err != nil {
			return errors.FromError(err)
		}
	}

	err = dbGameToProto(dg, out)
	if err != nil {
		return errors.FromError(err)
	}

	return nil
}

func (h *Handler) Update(ctx context.Context, in *gamedbpb.Game, out *gamedbpb.Game) error {
	dg := &db.Game{}
	err := protoGameToDB(in, dg)
	if err != nil {
		return errors.FromError(err)
	}

	// Get the database engine
	bun, err := component.Bun(h.cRegistry)
	if err != nil {
		return errors.FromError(err)
	}

	var result db.Game
	err = bun.NewSelect().
		Model(&result).
		Relation("Players").
		Limit(1).
		Where("g.id = ?", dg.Id).Scan(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	if err := h.checkGame(ctx, dg, &result); err != nil {
		return err
	}

	// Finaly update
	_, err = bun.NewUpdate().Model(dg).Exec(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	err = dbGameToProto(dg, out)
	if err != nil {
		return errors.FromError(err)
	}

	return nil
}

func (h *Handler) Delete(ctx context.Context, in *gamedbpb.DeleteRequest, out *empty.Empty) error {
	user, err := utils.CtxMetadataUser(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	if auth.IntersectsRoles(user, auth.AllowServiceAndAdmin...) {
		return errors.New(h.svcName, "Your not allowed to make that request", http.StatusMethodNotAllowed)
	}

	// Get the database engine
	bun, err := component.Bun(h.cRegistry)
	if err != nil {
		return err
	}

	// Execute the Delete
	_, err = bun.NewDelete().Model((*db.Game)(nil)).Where("g.id = ?", in.Id).Exec(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	return nil
}
