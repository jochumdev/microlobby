package gamedb

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4/errors"
	"go-micro.dev/v4/server"
	"jochum.dev/jo-micro/auth2"
	"jochum.dev/jo-micro/buncomponent"
	"jochum.dev/jo-micro/components"
	"jochum.dev/jo-micro/router"
	"wz2100.net/microlobby/service/gamedb/v1/db"
	"wz2100.net/microlobby/shared/proto/gamedbpb/v1"
)

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

const Name = "gamedbHandler"

type Handler struct {
	cReg        *components.Registry
	initialized bool
}

func New() *Handler {
	return &Handler{initialized: false}
}

func MustReg(cReg *components.Registry) *Handler {
	return cReg.Must(Name).(*Handler)
}

func (h *Handler) Name() string {
	return Name
}

func (h *Handler) Priority() int {
	return 100
}

func (h *Handler) Initialized() bool {
	return h.initialized
}

func (h *Handler) Init(components *components.Registry, cli *cli.Context) error {
	if h.initialized {
		return nil
	}

	h.cReg = components

	r := router.MustReg(h.cReg)
	r.Add(
		router.NewRoute(
			router.Method(router.MethodGet),
			router.Path("/"),
			router.Endpoint(gamedbpb.GameDBV1Service.List),
			router.Params("id", "history", "name", "limit", "offset"),
			router.AuthRequired(),
		),
		router.NewRoute(
			router.Method(router.MethodPost),
			router.Path("/"),
			router.Endpoint(gamedbpb.GameDBV1Service.Create),
			router.AuthRequired(),
		),
		router.NewRoute(
			router.Method(router.MethodPut),
			router.Path("/:id"),
			router.Endpoint(gamedbpb.GameDBV1Service.Update),
			router.Params("id"),
			router.AuthRequired(),
		),
		router.NewRoute(
			router.Method(router.MethodDelete),
			router.Path("/:id"),
			router.Endpoint(gamedbpb.GameDBV1Service.Delete),
			router.Params("id"),
			router.AuthRequired(),
		),
	)

	gamedbpb.RegisterGameDBV1ServiceHandler(h.cReg.Service().Server(), h)

	h.initialized = true
	return nil
}

func (h *Handler) Stop() error {
	return nil
}

func (h *Handler) Flags(r *components.Registry) []cli.Flag {
	return []cli.Flag{}
}

func (h *Handler) Health(context context.Context) error {
	return nil
}

func (h *Handler) WrapHandlerFunc(ctx context.Context, req server.Request, rsp interface{}) error {
	return nil
}

func (h *Handler) List(ctx context.Context, in *gamedbpb.ListRequest, out *gamedbpb.ListResponse) error {
	user, err := auth2.ClientAuthMustReg(h.cReg).Plugin().Inspect(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	if in.History {
		if !auth2.IntersectsRoles(user, auth2.RolesServiceAndAdmin...) {
			return errors.BadRequest("NOT_ALLOWED", "Your not allowed to make history requests")
		}
	}

	count, err := buncomponent.MustReg(h.cReg).Bun().NewSelect().Model((*db.Game)(nil)).Count(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	var games []db.Game
	err = buncomponent.MustReg(h.cReg).Bun().NewSelect().
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
	user, err := auth2.ClientAuthMustReg(h.cReg).Plugin().Inspect(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	if dg.V3GameId != 0 {
		if !auth2.IntersectsRoles(user, auth2.RolesServiceAndAdmin...) {
			return errors.BadRequest("NOT_ALLOWED", "You'r not allowed to make V3GameID requests")
		}

		if dg.LobbyVersion != 3 {
			return errors.BadRequest("NOT_ALLOWED_V3_GAME", "V3GameID is only allowed for LobbyVersion 3")
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
		return errors.BadRequest("NO_MATCH", "No host player given or IP's don't match")
	}

	if !auth2.IntersectsRoles(user, auth2.RolesServiceAndAdmin...) && nHostPlayer.UUID.String() != user.Id {
		return errors.BadRequest("NOT_ALLOWED", "Your only allowed to host own games")
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
			return errors.MethodNotAllowed("UPDATE_NOT_ALLOWED", "An update is not allowed")
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

	_, err = buncomponent.MustReg(h.cReg).Bun().NewInsert().
		Model(dg).
		Exec(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	for _, dp := range dg.Players {
		dp.GameID = dg.Id
		_, err = buncomponent.MustReg(h.cReg).Bun().NewInsert().
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

	var result db.Game
	err = buncomponent.MustReg(h.cReg).Bun().NewSelect().
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
	_, err = buncomponent.MustReg(h.cReg).Bun().NewUpdate().Model(dg).Exec(ctx)
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
	user, err := auth2.ClientAuthMustReg(h.cReg).Plugin().Inspect(ctx)
	if err != nil {
		return errors.FromError(err)
	}
	if !auth2.IntersectsRoles(user, auth2.RolesServiceAndAdmin...) {
		var dg db.Game
		err = buncomponent.MustReg(h.cReg).Bun().NewSelect().
			Model(&dg).
			Relation("Players").
			Limit(1).
			Where("g.id = ?", in.Id).Scan(ctx)
		if err != nil {
			return errors.FromError(err)
		}

		var nHostPlayer *db.GamePlayer
		for _, dp := range dg.Players {
			if dp.IsHost && dp.IpAddress == dg.HostIp {
				nHostPlayer = dp
				break
			}
		}
		if nHostPlayer == nil {
			return errors.BadRequest("NO_MATCH", "No host player given or IP's don't match")
		}

		if nHostPlayer.UUID.String() != user.Id {
			return errors.BadRequest("NOT_ALLOWED", "Your not allowed to delete foreign games")
		}
	}

	// Execute the Delete
	_, err = buncomponent.MustReg(h.cReg).Bun().NewDelete().Model((*db.Game)(nil)).Where("g.id = ?", in.Id).Exec(ctx)
	if err != nil {
		return errors.FromError(err)
	}

	return nil
}
